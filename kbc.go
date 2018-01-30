package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrNoMatch = errors.New("nope")
	re         = regexp.MustCompile(`\s+(\d\d [A-Z][a-z]{2} 201\d)\s+(.*)\s\s+([,0-9]+\.\d+)\s\s+([,0-9]+\.\d+)`)
)

type row struct {
	date    time.Time
	item    string
	change  decimal.Decimal
	diff    decimal.Decimal
	balance decimal.Decimal
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
		log.Print(err)
		return out, ErrNoMatch
	}
	out.date = d
	out.item = strings.TrimSpace(item)
	c, err := decomma(change)
	if err != nil {
		log.Print(err)
		return out, ErrNoMatch
	}
	out.change = c
	b, err := decomma(balance)
	if err != nil {
		log.Print(err)
		return out, ErrNoMatch
	}
	out.balance = b
	return out, nil
}

func decomma(in string) (decimal.Decimal, error) {
	x := strings.Replace(in, ",", "", -1)
	return decimal.NewFromString(x)
}

func parseDoc(fd io.Reader, rejects io.Writer) ([]row, error) {
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

func main() {
	f := "/Users/psn/Downloads/wat/Current Account Statement - 01 Oct 2017.txt"
	fd, err := os.Open(f)
	defer fd.Close()
	if err != nil {
		log.Fatal(err)
	}
	rejects, err := os.Create("rejects")
	defer rejects.Close()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := parseDoc(fd, rejects)
	if err != nil {
		log.Fatal(err)
	}
	var balance decimal.Decimal
	for i, r := range rows {
		if i == 0 {
			balance = r.balance
			fmt.Println(r)
			continue
		}
		diff := r.balance.Sub(balance)
		if i < 10 {
			fmt.Println(r.item, diff, r.balance)
		}

		invert := r.change.Mul(decimal.NewFromFloat(-1))
		if !diff.Equal(r.change) && !diff.Equal(invert) {
			log.Fatal(r, "s", diff, r.change, r.change.Mul(decimal.NewFromFloat(-1)))
		}
		r.diff = diff
		balance = r.balance
	}
}
