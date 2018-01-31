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
	id, err := newSheet(ctx, srv, sheet, "test")
	if err != nil {
		return err
	}
	req := &sheets.BatchUpdateValuesRequest{}
	vr := &sheets.ValueRange{}
	vr.MajorDimension = "ROWS"
	headings := []interface{}{
		"date",
		"description",
		"credit",
		"debit",
		"balance",
		"bucket",
	}
	vr.Values = append(vr.Values, headings)
	for i, r := range rows {
		if i > 10 {
			break
		}
		var credit, debit string
		if r.diff.LessThan(decimal.Decimal{}) {
			credit = ""
			debit = r.diff.StringFixed(2)
		} else {
			debit = ""
			credit = r.diff.StringFixed(2)
		}

		rr := []interface{}{
			r.date.Format("2006/01/02/"),
			r.item,
			credit,
			debit,
			r.balance.StringFixed(2),
			r.class,
		}
		vr.Values = append(vr.Values, rr)
	}
	req.Data = append(req.Data, vr)

	_, err = srv.Spreadsheets.BatchUpdate(sheet, req).Context(ctx).Do()
	return err
}

func uploadOneFileOld(ctx context.Context, srv *sheets.Service, sheet string, rows []row, name string) error {
	id, err := newSheet(ctx, srv, sheet, "test")
	if err != nil {
		return err
	}
	requests := []*sheets.Request{}
	req := &sheets.Request{}
	var rowData []*sheets.RowData
	rr := &sheets.RowData{}
	cd := &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: "date",
		},
	}
	rr.Values = append(rr.Values, cd)
	cd = &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: "Description",
		},
	}
	rr.Values = append(rr.Values, cd)
	cd = &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: "Change",
		},
	}
	rr.Values = append(rr.Values, cd)
	cd = &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: "Balance",
		},
	}
	rr.Values = append(rr.Values, cd)
	cd = &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: "Bucket",
		},
	}
	rr.Values = append(rr.Values, cd)
	rowData = append(rowData, rr)

	for i, r := range rows {
		if i > 10 {
			break
		}
		rr := &sheets.RowData{}
		cd := &sheets.CellData{
			UserEnteredValue: &sheets.ExtendedValue{
				StringValue: r.date.Format("02/01/2006"),
			},
			UserEnteredFormat: &sheets.CellFormat{
				NumberFormat: &sheets.NumberFormat{
					Type:    "DATE",
					Pattern: "dd/mm/yyyy",
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

		cd = &sheets.CellData{
			UserEnteredFormat: &sheets.CellFormat{
				NumberFormat: &sheets.NumberFormat{
					Type: "CURRENCY",
				},
			},
			UserEnteredValue: &sheets.ExtendedValue{StringValue: r.balance.StringFixed(2)},
		}
		rr.Values = append(rr.Values, cd)
		cd = &sheets.CellData{
			UserEnteredValue: &sheets.ExtendedValue{StringValue: r.class},
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
