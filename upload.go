package main

import (
	"golang.org/x/net/context"
	sheets "google.golang.org/api/sheets/v4"
)

func newSheet(ctx context.Context, srv *sheets.Service, sheet string, name string) (int64, error) {
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
			UserEnteredValue: &sheets.ExtendedValue{StringValue: r.item},
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
