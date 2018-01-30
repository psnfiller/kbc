package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrNoMatch   = errors.New("nope")
	ErrFloat     = errors.New("failed to parse float")
	re           = regexp.MustCompile(`\s+(\d\d [A-Z][a-z]{2} 201\d)\s+(.*)\s\s+([,0-9]+\.\d+)\s\s+([,0-9]+\.\d+)`)
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
		log.Print(err)
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

func classify(in string) string {
	switch {
	case in == "SDD KBC Bank Ireland Public Limited":
		return "House"
	case strings.Contains(in, " PROPERTY TAX"):
		return "House"
	case strings.Contains(in, "POS THORNTONS RECYCLING"):
		return "House"
	case strings.Contains(in, "SDD ELECTRIC IRELAND"):
		return "House"
	case strings.Contains(in, "Flogas Natural Gas"):
		return "House"
	case strings.Contains(in, "EIRCOM"):
		return "House"

	case strings.Contains(in, "POS TESCO STORE"):
		return "Food"
	case strings.Contains(in, "POS FARRELLYS"):
		return "Food"
	case strings.Contains(in, "POS THE DELGANY GROCER"):
		return "Food"
	case strings.Contains(in, "POS WWW.BRITISHFINEFOODS"):
		return "Food"
	case strings.Contains(in, "POS SHERIDANS CHEESEMONG"):
		return "Food"
	case strings.Contains(in, "FIREHOUSE BAKERY"):
		return "Food"
	case strings.Contains(in, "MINI MARKET DELGANY"):
		return "Food"
	case strings.Contains(in, "THE BEAR PAW DELI"):
		return "Food"
	case strings.Contains(in, "FX BUCKLEY BUTCHERS"):
		return "Food"

	case strings.Contains(in, "THE TRAMYARD KITCHEN"):
		return "meals out"

	case strings.Contains(in, " Google Ireland Limited"):
		return "Pay"

	case strings.Contains(in, "Smart Move Online"):
		return "Savings"
	case strings.Contains(in, "SCT Peter Nuttall & Laura Nuttall"):
		return "Savings"
	case strings.Contains(in, "STO SCT Peter Nuttall"):
		return "Savings"
	case strings.Contains(in, "STO SCT Peter and Laura Nuttall"):
		return "Savings"
	case strings.Contains(in, "STO SCT Laura Nuttall"):
		return "Savings"
	case strings.Contains(in, "KBC Mobile : Extra Regular Saver"):
		return "Savings"

	case strings.Contains(in, "PAYPAL COMPULABLTD"):
		return "Computers"
	case strings.Contains(in, "APPLE ONLINE STORE"):
		return "Computers"
	case strings.Contains(in, "GOOGLE CLOUD"):
		return "Computers"
	case strings.Contains(in, "LASTPASS.COM"):
		return "Computers"
	case strings.Contains(in, "GITHUB.COM"):
		return "Computers"

	case strings.HasPrefix(in, "ATM"):
		return "Cash"

	case strings.HasPrefix(in, "POS Amazon"):
		return "Amazon"
	case strings.HasPrefix(in, "POS AMAZON.CO.UK"):
		return "Amazon"
	case strings.HasPrefix(in, "POS AMAZON.UK"):
		return "Amazon"
	case strings.HasPrefix(in, "POS AMAZON SVCS EU-UK AM"):
		return "Amazon"
	case strings.HasPrefix(in, "AMAZON MKTPLACE PMTS"):
		return "Amazon"
	case strings.HasPrefix(in, "AMAZON EU"):
		return "Amazon"

	case strings.Contains(in, "VODAFONE"):
		return "phone"

	case strings.Contains(in, "LLOYDSPHARMACY"):
		return "medical"
	case strings.Contains(in, "Blackrock Clinic Spe"):
		return "medical"
	case strings.Contains(in, "McGleenans Pharmacy"):
		return "medical"

	case strings.Contains(in, "DUBRAY BOOKS"):
		return "books"
	case strings.Contains(in, "The Village Bookshop"):
		return "books"
	case strings.Contains(in, "WATERSTONES"):
		return "books"

	default:
		return defaultClass
	}
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
			rejects.Write([]byte(line + "\n"))
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
	var sum decimal.Decimal
	var classified decimal.Decimal
	buckets := make(map[string]decimal.Decimal)
	for i, r := range rows {
		if i == 0 {
			balance = r.balance
			continue
		}
		diff := r.balance.Sub(balance)
		invert := r.change.Mul(decimal.NewFromFloat(-1))
		if !diff.Equal(r.change) && !diff.Equal(invert) {
			log.Fatal(r, "s", diff, r.change, r.change.Mul(decimal.NewFromFloat(-1)))
		}
		r.diff = diff
		balance = r.balance
	}
	sort.Slice(rows, func(a, b int) bool {
		return rows[a].diff.GreaterThan(rows[b].diff)
	})
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
		buckets[r.class] = b.Add(r.diff)
	}
	fmt.Printf("%s %s %s%%\n", sum, classified, classified.Mul(decimal.NewFromFloat(100)).DivRound(sum, 2))
	for b, d := range buckets {
		fmt.Println(b, d)
	}
}
