# receipt-processor-api
This is a contrived API and an implementation of [the receipt processor challenge](https://github.com/fetch-rewards/receipt-processor-challenge). See [this reddit thread](https://www.reddit.com/r/golang/comments/1dtvolz/feedback_after_being_rejected_for_my_take_home/) for more info.

## Notes about this project / feedback
- Due to time constraints, not a lot of validation is being done on data posted.
- It's odd that numbers are being represented in incoming JSON as strings. I probably would have implemented a custom [Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal) to normalize the data if I had more time to keep my downstream logic more pristine.
- I found it odd that there was a total at the receipt level when it could be calculated by the items. Perhaps an attempt to get developers to make note of that?
- I have been writing software professionally for almost 20 years and I have yet to see a requirement like "If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer". I understand the rationale, but, it would have been more fun if they had attempted to find real-world stuff like "if the receipt originated in WA, apply a tax rate of X to the total".
- Some things were (probably intentionally) ambiguous. Like "If the trimmed length of `some string`...". Length in what sense? String length is complicated. From [the docs](https://go.dev/blog/strings): "But what about the lower case grave-accented letter ‘A’, à? That’s a character, and it’s also a code point (U+00E0), but it has other representations. For example we can use the “combining” grave accent code point, U+0300, and attach it to the lower case letter a, U+0061, to create the same character à".
- I clearly decided to go the barebones route. A single file and very few dependencies. I think this is [inline with idiomatic code for small projects](https://go.dev/doc/modules/layout#basic-command).

## Running the project
Make sure you have [Go installed](https://go.dev/dl/). Then run the following commands:

```bash
# First, download dependencies
go mod tidy
# Next, run the project.
go run main.go
```

To test, you can use curl to post data:

```bash
curl -d '{ "retailer": "M&M Corner Market", "purchaseDate": "2022-03-20", "purchaseTime": "14:33", "items": [ { "shortDescription": "Gatorade", "price": "2.25" },{ "shortDescription": "Gatorade", "price": "2.25" },{ "shortDescription": "Gatorade", "price": "2.25" },{ "shortDescription": "Gatorade", "price": "2.25" } ], "total": "9.00" }' -H "Content-Type: application/json" -X POST http://localhost:8080/receipts/process
```

Take note of the returned ID. You can either just open `http://localhost:8080/receipts/{the ID you recorded}/points` in your browser, or use another curl command to check it:

```bash
curl localhost:8080/receipts/{the ID you recorded}/points
```

## Testing the project
Run `go test` to run unit tests.