package main

import (
	"log"

	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
	sheets "google.golang.org/api/sheets/v4"
)

func newSheet(ctx context.Context, srv *sheets.Service, sheet string, name string) (int64, error) {
	sheetResp, err := srv.Spreadsheets.Get(sheet).Context(ctx).Do()
	if err != nil {
		return 0, err
	}
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

func uploadOneFile(ctx context.Context, srv *sheets.Service, sheet string, rows []row, name string) error {
	_, err := newSheet(ctx, srv, sheet, name)
	if err != nil {
		return err
	}
	req := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data: []*sheets.ValueRange{&sheets.ValueRange{
			MajorDimension: "ROWS",
			Values: [][]interface{}{
				{
					"date",
					"description",
					"credit",
					"debit",
					"balance",
					"bucket",
				}},
		}},
	}
	vr := &sheets.ValueRange{MajorDimension: "ROWS"}
	headings := []interface{}{}
	vr.Values = append(vr.Values, headings)
	for _, r := range rows {
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
	vr.Range = "name"
	req.Data = append(req.Data, vr)

	ss := sheets.NewSpreadsheetsValuesService(srv)
	_, err = ss.BatchUpdate(sheet, req).Context(ctx).Do()
	return err
}
