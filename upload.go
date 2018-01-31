package main

import (
	"log"

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

	// Does the sheet have a tab?
	requests := []*sheets.Request{}
	req := &sheets.Request{}
	req.AddSheet = &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title: name,
		},
	}
	requests = append(requests, req)
	rb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}
	resp, err := srv.Spreadsheets.BatchUpdate(sheet, rb).Context(ctx).Do()
	if err != nil {
		return 0, err
	}
	return resp.Replies[0].AddSheet.Properties.SheetId, nil
}

func uploadOneFile(ctx context.Context, srv *sheets.Service, sheet string, rows []row, name string) error {
	id, err := newSheet(ctx, srv, sheet, name)
	if err != nil {
		return err
	}
	requests := []*sheets.Request{}
	req := &sheets.Request{}
	var rowData []*sheets.RowData
	for _, r := range rows {
		rr := &sheets.RowData{}
		cd := &sheets.CellData{
			UserEnteredValue: &sheets.ExtendedValue{
				StringValue: r.date.Format("2006-01-02"),
			},
			UserEnteredFormat: &sheets.CellFormat{
				NumberFormat: &sheets.NumberFormat{
					Type:    "DATE",
					Pattern: "yyyy-mm-dd",
				},
			},
		}

		rr.Values = append(rr.Values, cd)
		cd = &sheets.CellData{
			UserEnteredValue: &sheets.ExtendedValue{StringValue: r.item},
		}
		rr.Values = append(rr.Values, cd)

		cd = &sheets.CellData{
			UserEnteredFormat: &sheets.CellFormat{
				NumberFormat: &sheets.NumberFormat{
					Type: "CURRENCY",
				},
			},
			UserEnteredValue: &sheets.ExtendedValue{StringValue: r.diff.StringFixed(2)},
		}
		rr.Values = append(rr.Values, cd)

		rowData = append(rowData, rr)
	}
	req.AppendCells = &sheets.AppendCellsRequest{
		SheetId: id,
		Rows:    rowData,
		Fields:  "*",
	}
	requests = append(requests, req)
	rb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}
	_, err = srv.Spreadsheets.BatchUpdate(sheet, rb).Context(ctx).Do()
	return err
}
