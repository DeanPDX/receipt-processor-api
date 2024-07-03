package main

import (
	"encoding/json"
	"testing"
)

func Test_countAlphanumeric(t *testing.T) {
	tests := []struct {
		testName  string
		testVal   string
		wantCount int
	}{
		{"Simple string", "shouldbefine", 12},
		{"Ignore whitespace and periods", "This is a\n test.", 11},
		{"Handle extended characters", "日本語", 0},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := countAlphanumeric(tt.testVal); got != tt.wantCount {
				t.Errorf("countAlphanumeric(\"%v\") = %v, want %v", tt.testVal, got, tt.wantCount)
			}
		})
	}
}

func Test_receipt_points(t *testing.T) {
	// This JSON and corresponding expected result is taken from:
	// https://github.com/fetch-rewards/receipt-processor-challenge?tab=readme-ov-file#examples
	firstTestJSON := `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },{
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    },{
      "shortDescription": "Knorr Creamy Chicken",
      "price": "1.26"
    },{
      "shortDescription": "Doritos Nacho Cheese",
      "price": "3.35"
    },{
      "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
      "price": "12.00"
    }
  ],
  "total": "35.35"
}`
	var testReceipt receipt
	err := json.Unmarshal([]byte(firstTestJSON), &testReceipt)
	if err != nil {
		t.Errorf("Wasn't expecting error. Got: %v", err)
	}
	/*
		Total Points: 28
		Breakdown:
			6 points - retailer name has 6 characters
			10 points - 4 items (2 pairs @ 5 points each)
			3 Points - "Emils Cheese Pizza" is 18 characters (a multiple of 3)
						item price of 12.25 * 0.2 = 2.45, rounded up is 3 points
			3 Points - "Klarbrunn 12-PK 12 FL OZ" is 24 characters (a multiple of 3)
						item price of 12.00 * 0.2 = 2.4, rounded up is 3 points
			6 points - purchase day is odd
	*/
	want := 28
	got := testReceipt.points()
	if want != got {
		t.Errorf("Want %v. Got %v.", want, got)
	}

	secondTestJSON := `{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    }
  ],
  "total": "9.00"
}`
	err = json.Unmarshal([]byte(secondTestJSON), &testReceipt)
	if err != nil {
		t.Errorf("Wasn't expecting error. Got: %v", err)
	}
	/*
		Total Points: 109
		Breakdown:
			50 points - total is a round dollar amount
			25 points - total is a multiple of 0.25
			14 points - retailer name (M&M Corner Market) has 14 alphanumeric characters
						note: '&' is not alphanumeric
			10 points - 2:33pm is between 2:00pm and 4:00pm
			10 points - 4 items (2 pairs @ 5 points each)
	*/
	want = 109
	got = testReceipt.points()
	if want != got {
		t.Errorf("Want %v. Got %v.", want, got)
	}
}
