// Program KBC converts kbc's PDFs into google spreadsheets, and attempts to classify spending.
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
)

var (
	ErrNoMatch = errors.New("Failed to match line.")
	ErrFloat   = errors.New("failed to parse float")
	ErrTime    = errors.New("failed to parse Time")

	//03 Apr 2017     POS GITHUB.COM DFMK 20170329                    6.51                                  3,947.28
	// Deposit statements look different:
	// 01/01/2015       Opening Balance
	re = regexp.MustCompile(`\f?\s*(\d\d[/ ]([A-Z][a-z]{2}|\d+)[/ ]201\d)\s+(.*)\s\s+([,0-9]+\.\d+)\s\s+([,0-9]+\.\d+)`)

	defaultOut   = "unknown"
	defaultClass = "in unknown"

	sheetID      = flag.String("spreadsheet_id", "", "Id of the google spreadsheet to update")
	directory    = flag.String("directory", "", "Directory of files to upload")
	rejectsFile  = flag.String("rejects", "", "File to write unmatched lines to. If empty, then no file is used.")
	printBuckets = flag.Bool("buckets", false, "If true, print out spending per bucket and top unclassified expenses.")
	bucketCutOff = flag.Float64("bucket-cut-off", 5., "only print expenses above this value")
	csvFile      = flag.String("csv", "", "CSV output file.")
)

func validateFlags() error {
	if *directory == "" {
		return errors.New("you need to provide a directory, with --directory")
	}
	return nil
}

type row struct {
	// The date the transaction was made, from the statement.
	date time.Time
	// Description of the transaction.
	description string
	// Recorded change in the amount in the account. This can be either a debit or a credit, but is always >0. This is the result of reading the statement via dodgy regexps.
	change decimal.Decimal
	//
	diff decimal.Decimal
	// The balance in the account after the txn.
	balance decimal.Decimal
	class   string
}

func (r row) String() string {
	return fmt.Sprintf("%s\t%s\t%v\t%v", r.date.Format("2006-01-02"), r.description, r.change, r.balance)
}

// parseLine converts a line into a row, or an error if it can't parse the line.
func parseLine(line string) (row, error) {
	out := row{}
	l := re.FindStringSubmatch(line)
	if l == nil {
		return out, ErrNoMatch
	}
	date := l[1]
	// TODO(psn):
	description := l[3]
	change := l[4]
	balance := l[5]
	d, err := time.Parse("02 Jan 2006", date)
	if err != nil {
		// Try the other format
		d, err = time.Parse("02/01/2006", date)
		if err != nil {
			return out, ErrTime
		}
	}
	out.date = d
	out.description = strings.TrimSpace(description)
	c, err := decomma(change)
	if err != nil {
		log.Print(err)
		return out, ErrFloat
	}
	out.change = c
	b, err := decomma(balance)
	if err != nil {
		log.Print(err)
		return out, ErrFloat
	}
	out.balance = b
	out.class = classify(out.description, out.change)
	return out, nil
}

// decomma parses strings like "1,234.44"
func decomma(in string) (decimal.Decimal, error) {
	x := strings.Replace(in, ",", "", -1)
	return decimal.NewFromString(x)
}

// parseDoc reads all the lines from fd, and writes lines that don't match the regexp into rejects, if it is non-nil.
func parseDoc(fd io.Reader, rejects io.Writer) ([]row, error) {
	var out []row
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		r, err := parseLine(line)
		if err == ErrNoMatch {
			if rejects != nil {
				rejects.Write([]byte(line + "\n"))
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func processOneFile(filename string, rejects io.Writer) ([]row, error) {
	var out []row
	fd, err := os.Open(filename)
	defer fd.Close()
	if err != nil {
		return out, err
	}
	rows, err := parseDoc(fd, rejects)
	if err != nil {
		return out, err
	}
	// Check that for each row, the difference + the previous balance equals the current balance. This also enables working out of a difference is a credit or a debit.
	var balance decimal.Decimal
	for i, r := range rows {
		if i == 0 {
			balance = r.balance
			continue
		}
		diff := r.balance.Sub(balance)
		invert := r.change.Mul(decimal.NewFromFloat(-1))
		if !diff.Equal(r.change) && !diff.Equal(invert) {
			return rows, fmt.Errorf("failed to do diff %s %s %s %s", r, diff, r.change, invert)
		}
		rows[i].diff = diff
		if diff.IsNegative() && rows[i].class == defaultClass {
			rows[i].class = defaultOut
		}
		balance = r.balance
	}
	return rows, nil
}

func main() {
	flag.Parse()
	if err := validateFlags(); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	srv, err := newSrv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var rejects io.Writer
	if *rejectsFile != "" {
		rejects, err = os.Create(*rejectsFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	contents, err := ioutil.ReadDir(*directory)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cloud")
	var rows []row
	for _, c := range contents {
		if !strings.HasSuffix(c.Name(), ".txt") {
			continue
		}
		p := path.Join(*directory, c.Name())
		fmt.Println("cloud")
		r, err := processOneFile(p, rejects)
		if err != nil {
			log.Fatalf("failed to process %s: %s", c.Name(), err)
		}
		fmt.Println("cloud")
		rows = append(rows, r...)
		if *sheetID != "" {
			err = uploadOneFile(ctx, srv, *sheetID, rows, c.Name())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Println("cloud")
	if *sheetID != "" {
		err = uploadOneFile(ctx, srv, *sheetID, rows, "all")
		if err != nil {
			log.Fatal(err)
		}
	}
	if *printBuckets {
		buckets(rows)
	}
	if *csvFile != "" {
		fd, err := os.Create(*csvFile)
		if err != nil {
			log.Fatal(err)
		}
		if err := csvExport(rows, fd); err != nil {
			log.Fatal(err)
		}
	}
}

func csvExport(rows []row, fd io.Writer) error {
	ww := csv.NewWriter(fd)
	if err := ww.Write([]string{
		"date",
		"description",
		"diff",
		"balance",
		"class",
	}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := ww.Write([]string{
			r.date.Format("2006-01-02"),
			r.description,
			r.diff.StringFixed(2),
			r.balance.StringFixed(2),
			r.class}); err != nil {
			return err
		}
	}
	ww.Flush()
	return ww.Error()
}

func buckets(rows []row) {
	var sum decimal.Decimal
	var classified decimal.Decimal
	// Group the expenses into buckets.
	buckets := make(map[string]decimal.Decimal)
	for _, r := range rows {
		if r.change.GreaterThan(decimal.NewFromFloat(*bucketCutOff)) && r.class == defaultOut {
			// Print items in the default bucket, if greater than cutoff.
			fmt.Println(r)
		}
		sum = sum.Add(r.change)
		if r.class != defaultOut || r.class != defaultClass {
			classified = classified.Add(r.change)
		}
		b := buckets[r.class]
		buckets[r.class] = b.Add(r.diff)
	}
	// Print percentage classified.
	fmt.Printf("%f%% classified\n", classified.Mul(decimal.NewFromFloat(100)).DivRound(sum, 2))
	// Print out buckets.
	type bucks struct {
		name  string
		value decimal.Decimal
	}
	var b []bucks
	for bb, d := range buckets {
		b = append(b, bucks{bb, d})
	}
	sort.Slice(b, func(a, bb int) bool {
		return b[a].value.LessThan(b[bb].value)
	})
	for _, x := range b {
		fmt.Printf("%s\t%s\n", x.name, x.value)
	}
}
