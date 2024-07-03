/*
 Receipt-processor-api is a demo implementation of the following challenge:
 https://github.com/fetch-rewards/receipt-processor-challenge

 For more rationale on where this came from, see the following reddit post:
 https://www.reddit.com/r/golang/comments/1dtvolz/feedback_after_being_rejected_for_my_take_home/
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// countAlphanumeric will count the number of alphanumeric runes in the given string
// as defined in `const alphanumeric`.
func countAlphanumeric(s string) int {
	count := 0
	// Be careful when modifying this code to understand the difference between characters and code points:
	// https://go.dev/blog/strings#code-points-characters-and-runes
	for _, v := range s {
		if strings.ContainsRune(alphanumeric, v) {
			count++
		}
	}
	return count
}

// receipt is a struct representing the receipt object:
// https://github.com/fetch-rewards/receipt-processor-challenge/tree/main/examples
type receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []item `json:"items"`
	Total        string `json:"total"`
}

// item represents an item on a receipt:
// https://github.com/fetch-rewards/receipt-processor-challenge/blob/main/examples/morning-receipt.json#L7
type item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// points will tally the points for a receipt. Points are given with the following rules:
//   - One point for every alphanumeric character in the retailer name.
//   - 50 points if the total is a round dollar amount with no cents.
//   - 25 points if the total is a multiple of `0.25`.
//   - 5 points for every two items on the receipt.
//   - If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
//   - 6 points if the day in the purchase day is odd.
//   - 10 points if the time of purchase is after 2:00pm and before 4:00pm.
func (r receipt) points() int {
	points := 0
	// Add a point for every alphanumeric character in retailer name.
	points += countAlphanumeric(r.Retailer)
	// Add 50 points if total is integer
	// Ignoring error here (and elsewhere in this function)
	// because an invalid string is the same as a zero value in effect.
	total, _ := strconv.ParseFloat(r.Total, 64)
	if total == math.Trunc(total) {
		points += 50
	}
	// 25 points if the total is a multiple of `0.25`.
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}
	// 5 points for every two items on the receipt.
	totalLines := len(r.Items)
	if totalLines > 2 {
		points += totalLines / 2 * 5
	}
	for _, item := range r.Items {
		// If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer.
		// "length" is poorly defined here. I am assuming number of runes is the correct interpretation.
		descLen := utf8.RuneCountInString(strings.TrimSpace(item.ShortDescription))
		if descLen%3 == 0 {
			itemPrice, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(itemPrice * 0.2))
		}
	}
	// Combine date and time into a parsed time.Time
	purchasedAt, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%v %v", r.PurchaseDate, r.PurchaseTime))
	// If we have an error, we can just bail early.
	if err != nil {
		return points
	}
	// 6 points if the day in the purchase date is odd.
	if purchasedAt.Day()%2 != 0 {
		points += 6
	}
	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	// "After" 2PM could mean literally that (so we could check for 2:01PM)
	// but I believe the
	if purchasedAt.Hour() >= 14 && purchasedAt.Hour() <= 16 {
		points += 10
	}
	return points
}

// Global var for storing receipts in memory
var receipts = sync.Map{}

type processReceiptResp struct {
	ID string `json:"id"`
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	var rec receipt
	// Using json.NewDecoder will ignore malformed JSON after first
	// successful call to Decode:
	// https://github.com/golang/go/issues/36225
	// This is fine for our purposes but something to keep in mind
	// before deploying this app anywhere (though often it is the desired
	// result).
	err := json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, http.StatusText(http.StatusBadRequest))
		return
	}
	// Generate a new uuid and store the receipt object in our in-memory cache.
	id := uuid.New().String()
	receipts.Store(id, rec)
	resp := processReceiptResp{
		ID: id,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
	}
}

type pointsResp struct {
	Points int `json:"points"`
}

func receiptPoints(w http.ResponseWriter, r *http.Request) {
	value, ok := receipts.Load(r.PathValue("id"))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, http.StatusText(http.StatusNotFound))
		return
	}
	resp := pointsResp{
		Points: value.(receipt).points(),
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
	}
}

func main() {
	// It's probably not very REST-ful to have a verb ("process") in our URL.
	// This would make more sense as just a POST to /receipts since it's more
	// akin to a create. Following requirements over REST-ful API standards.
	http.HandleFunc("POST /receipts/process", processReceipt)
	http.HandleFunc("GET /receipts/{id}/points", receiptPoints)
	// TODO: before deploying anywhere, make this config or ENV-based.
	port := "8080"
	fmt.Printf("Listening on port %v\n", port)
	// TODO: before deploying anywhere, ideally we gracefully shut
	// down and drain connections. See also:
	// https://pkg.go.dev/net/http#Server.Shutdown
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

}
