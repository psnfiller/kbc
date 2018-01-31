package main

import (
	"bufio"
	"errors"
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
	ErrNoMatch   = errors.New("nope")
	ErrFloat     = errors.New("failed to parse float")
	re           = regexp.MustCompile(`\f?\s*(\d\d [A-Z][a-z]{2} 201\d)\s+(.*)\s\s+([,0-9]+\.\d+)\s\s+([,0-9]+\.\d+)`)
	defaultClass = "unknown"
)

type row struct {
	date    time.Time
	item    string
	change  decimal.Decimal
	diff    decimal.Decimal
	balance decimal.Decimal
	class   string
}

func (r row) String() string {
	return fmt.Sprintf("%s\t%s\t%v\t%v", r.date.Format("2006-01-02"), r.item, r.change, r.balance)
}

func parseLine(line string) (row, error) {
	out := row{}
	l := re.FindStringSubmatch(line)
	if l == nil {
		return out, ErrNoMatch
	}
	date := l[1]
	item := l[2]
	change := l[3]
	balance := l[4]
	d, err := time.Parse("02 Jan 2006", date)
	if err != nil {
		log.Print(err, line, date)
		return out, ErrFloat
	}
	out.date = d
	out.item = strings.TrimSpace(item)
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
	out.class = classify(out.item)
	return out, nil
}

func decomma(in string) (decimal.Decimal, error) {
	x := strings.Replace(in, ",", "", -1)
	return decimal.NewFromString(x)
}

func parseDoc(fd io.Reader) ([]row, error) {
	var out []row
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		r, err := parseLine(line)
		if err == ErrNoMatch {
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

func processOneFile(filename string) ([]row, error) {
	var out []row
	fd, err := os.Open(filename)
	defer fd.Close()
	if err != nil {
		return out, err
	}
	rows, err := parseDoc(fd)
	if err != nil {
		return out, err
	}
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
		balance = r.balance
	}
	return rows, nil
}

var (
	sheetID = "14TEOrodK2WsY87Y4XVbYLarfRVDYUVyMa6p6GsRNsE4"
)

func main() {
	ctx := context.Background()
	srv, err := newSrv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//dir := "/Users/psn/Downloads/wat"
	dir := "/Users/psn/Documents/statements"
	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var rows []row
	for _, c := range contents {
		if !strings.HasSuffix(c.Name(), ".txt") {
			continue
		}
		p := path.Join(dir, c.Name())
		r, err := processOneFile(p)
		if err != nil {
			log.Fatalf("failed to process %s: %s", c.Name(), err)
		}
		rows = append(rows, r...)
		err = uploadOneFile(ctx, srv, sheetID, rows, c.Name())
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = uploadOneFile(ctx, srv, sheetID, rows, "all")
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(rows, func(a, b int) bool {
		return rows[a].diff.GreaterThan(rows[b].diff)
	})

	var sum decimal.Decimal
	var classified decimal.Decimal
	buckets := make(map[string]decimal.Decimal)
	for _, r := range rows {
		t := 5.
		diff := r.diff
		if (diff.GreaterThan(decimal.NewFromFloat(t)) || diff.LessThan(decimal.NewFromFloat(-t))) && r.class == defaultClass {
			fmt.Println(r)
		}
		sum = sum.Add(r.change)
		if r.class != defaultClass {
			classified = classified.Add(r.change)
		}
		b := buckets[r.class]
		buckets[r.class] = b.Add(diff)
	}
	fmt.Printf("%s %s %s%%\n", sum, classified, classified.Mul(decimal.NewFromFloat(100)).DivRound(sum, 2))
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
