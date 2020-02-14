package main

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
	sheets "google.golang.org/api/sheets/v4"
)

// newSheet adds a tab to the sheet with the given name, or returns the id of the existing tab with the given name.
func newSheet(ctx context.Context, srv *sheets.Service, sheet string, name string) (int64, error) {
	fmt.Println("newS")
	sheetResp, err := srv.Spreadsheets.Get(sheet).Context(ctx).Do()
	if err != nil {
		return 0, err
	}
	fmt.Println("newS")
	found := false
	var id int64
	for _, s := range sheetResp.Sheets {
		if s.Properties.Title == name {
			found = true
			id = s.Properties.SheetId
		}
	}
	if found {
		log.Print("found existing sheet")
		return id, nil
	}

	rb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			&sheets.Request{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: name,
					},
				},
			},
		},
	}
	resp, err := srv.Spreadsheets.BatchUpdate(sheet, rb).Context(ctx).Do()
	if err != nil {
		return 0, err
	}
	return resp.Replies[0].AddSheet.Properties.SheetId, nil
}

// uploadOneFile adds the rows in `rows` to the tab named in `name` to the sheet with id `sheet`. If the tab does not exist, it is created.
func uploadOneFile(ctx context.Context, srv *sheets.Service, sheet string, rows []row, name string) error {
	fmt.Println("uploadone")
	_, err := newSheet(ctx, srv, sheet, name)
	if err != nil {
		return err
	}
	fmt.Println("uploadone")
	vr := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          name}
	headings := []interface{}{
		"date",
		"description",
		"credit",
		"debit",
		"balance",
		"bucket",
	}
	vr.Values = append(vr.Values, headings)
	for _, r := range rows {
		// We input dates and money with currency symbols etc, similar to human entry.
		var credit, debit string
		if r.diff.LessThan(decimal.Decimal{}) {
			credit = ""
			debit = "-€" + r.change.StringFixed(2)
		} else {
			debit = ""
			credit = "€" + r.diff.StringFixed(2)
		}

		rr := []interface{}{
			r.date.Format("2006/01/02"),
			r.description,
			credit,
			debit,
			"€" + r.balance.StringFixed(2),
			r.class,
		}
		vr.Values = append(vr.Values, rr)
	}
	req := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             []*sheets.ValueRange{vr},
	}

	ss := sheets.NewSpreadsheetsValuesService(srv)
	_, err = ss.BatchUpdate(sheet, req).Context(ctx).Do()
	return err
}
