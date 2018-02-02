# KBC parser.

This code, along with pkdtotext, converts KBC's online statements from PDFs into Google spreadsheets.

```
for i in *pdf; do pdftotext -layout $i ; done
```


## Dependencies

go get -u google.golang.org/api/sheets/v4
go get -u golang.org/x/oauth2/...

## Correctness

 * To avoid floating point comedy, this uses the [decimal library](https://godoc.org/github.com/shopspring/decimal#Decimal.Mul).
 * The program matches lines in the statement with a regexp. You can examine lines not matched by the regexp by running with `--rejects=/tmp/file` and reading the file.
 * The Program checks that the balances match the credits and the debits.

## Refs

https://godoc.org/google.golang.org/api/sheets/v4

https://developers.google.com/sheets/api/guides/formats

