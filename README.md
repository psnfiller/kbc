# KBC parser.

This code, along with pkdtotext, converts KBC's online statements from PDFs into Google spreadsheets.


## Dependencies

```
% go get -u google.golang.org/api/sheets/v4
% go get -u golang.org/x/oauth2/...
% go get github.com/shopspring/decimal
```

## Running

1) Create `client_secret.json` as documented in [this tutorial](https://developers.google.com/sheets/api/quickstart/go)

2) Convert PDFs to text with `for i in *pdf; do pdftotext -layout $i ; done`.

3) Write a `classify` function. You can use `exampleClassify` in `example_classify.go` to get started.

4) `go build && ./kbc --directory ~/Documents/statements --spreadsheet_id $id --rejects /tmp/f1`. You can get the spreadsheet id from the URL.

## Correctness

 * To avoid floating point comedy, this uses the [decimal library](https://godoc.org/github.com/shopspring/decimal#Decimal.Mul).
 * The program matches lines in the statement with a regexp. You can examine lines not matched by the regexp by running with `--rejects=/tmp/file` and reading the file.
 * The Program checks that the balances match the credits and the debits.

## Spreadsheet notes

The program makes a new tab (aka sheet) in the spreadsheet with the same name as the input file. It also produces one tab with all rows from all files.

The program inserts rows into the spreadsheet as `USER_ENTERED`, and the spreadsheet attempts to work out if they are currency, dates, etc. It could use the UserFormat options instead.

## Security.

* The program caches a oauth token in `~/.credentials`.
* You should check the sharing settings on your google doc.

## Refs

https://developers.google.com/sheets/api/quickstart/go
https://godoc.org/google.golang.org/api/sheets/v4
https://developers.google.com/sheets/api/guides/formats

# Bucketing.

For each expense, I try and classify it into a "bucket", for example food, travel, house, etc. `classify()` does this. As my `classify()` leaks infomation about my spending habits, then I provided an example. You can use `--bucket` to help write your `classify()` - it prints out a) the percentage of credits / debits covered by `classify()` and the highest unclassified expenses.

# Notice.

I'm not a KBC employee and this is not KBC's code. There is no warranty.

# TODO(psn):

 * create new sheets by hand.
 * CSV /TSV output, to feed into [q](http://harelba.github.io/q/)
 * Real database support.
 * Support for turning off `~/.credentials`

