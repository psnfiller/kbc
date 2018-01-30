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
	case strings.Contains(in, "EIR"):
		return "House"
	case strings.Contains(in, "IRISH LIFE"):
		return "House"

	case strings.Contains(in, "BROWN THOMAS"):
		return "House stuff"

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
	case strings.Contains(in, "3FE"):
		return "Food"
	case strings.Contains(in, "LIDL"):
		return "Food"
	case strings.Contains(in, "PAYPAL HAS BEAN"):
		return "Food"

	case strings.Contains(in, "THE TRAMYARD KITCHEN"):
		return "meals out"
	case strings.Contains(in, "CHAKRA BY JAIPUR"):
		return "meals out"
	case strings.Contains(in, "The Pigeon House Del"):
		return "meals out"
	case strings.Contains(in, "THE WATERLOO BAR"):
		return "meals out"
	case strings.Contains(in, "Just Eat"):
		return "meals out"
	case strings.Contains(in, "MANGO TREE GREYSTONE"):
		return "meals out"
	case strings.Contains(in, "HORSE HOUND"):
		return "meals out"
	case strings.Contains(in, "BOMBAY P"):
		return "meals out"

	case strings.Contains(in, "O BRIENS WINES"):
		return "booze"

	case strings.Contains(in, "POS GREAT OUTDOORS"):
		return "clothes"
	case strings.Contains(in, "HOTMILK LINGERIE"):
		return "clothes"

	case strings.Contains(in, "Footprints Montessori"):
		return "childcare"

	case strings.Contains(in, "MOTHERCARE"):
		return "baby clothes"
	case strings.Contains(in, "JOJOMAMANBEBE.CO.U"):
		return "baby clothes"
	case strings.Contains(in, "JOJO MAMAN BEBE"):
		return "baby clothes"

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
	case strings.Contains(in, "PAYPAL TARSNAPCOM"):
		return "Computers"
	case strings.Contains(in, "PROJECT ALLOY"):
		return "Computers"
	case strings.Contains(in, "PAYPAL LETSENCRYPT"):
		return "Computers"
	case strings.Contains(in, "PAYPAL FREEPRESS"):
		return "Computers"
	case strings.Contains(in, "PAYPAL OPENBSDFOUN"):
		return "Computers"
	case strings.Contains(in, "ZAZZLE"):
		return "Computers"

	case strings.Contains(in, "GOOGLE Google Music"):
		return "Music"
	case strings.Contains(in, "ITUNES.COMBILL"):
		return "Music"

	case strings.Contains(in, "IRISH RAIL"):
		return "Dart"

	case strings.Contains(in, "THE HAPPY PEAR"):
		return "Coffee"
	case strings.Contains(in, "COFFEEANGEL"):
		return "Coffee"

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
	case strings.Contains(in, "AMAZON"):
		return "Amazon"

	case strings.Contains(in, "VODAFONE"):
		return "phone"

	case strings.Contains(in, "LLOYDSPHARMACY"):
		return "medical"
	case strings.Contains(in, "Blackrock Clinic Spe"):
		return "medical"
	case strings.Contains(in, "McGleenans Pharmacy"):
		return "medical"
	case strings.Contains(in, "MEDICARE"):
		return "medical"
	case strings.Contains(in, "HAPPYHEARTCOURSECOM"):
		return "medical"
	case strings.Contains(in, "GREYSTONES HARBOUR F"):
		return "medical"

	case strings.Contains(in, "DUBRAY BOOKS"):
		return "books"
	case strings.Contains(in, "The Village Bookshop"):
		return "books"
	case strings.Contains(in, "WATERSTONES"):
		return "books"
	case strings.Contains(in, "Bridge Street Books"):
		return "books"
	case strings.Contains(in, "EASONS"):
		return "books"

	case strings.Contains(in, "INTERFLORA"):
		return "flowers"
	case strings.Contains(in, "COTTAGE FLOWER"):
		return "flowers"

	case strings.Contains(in, "AN POST-OFFICES"):
		return "flowers"

	case strings.Contains(in, "APPLEGREEN MOTORWAY"):
		return "fuel"

	case strings.Contains(in, "PARCEL MOTEL"):
		return "parcel motel"
	case strings.Contains(in, "PARCELMOTEL"):
		return "parcel motel"

	default:
		return defaultClass
	}
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
			log.Fatal(r, "s", diff, r.change, r.change.Mul(decimal.NewFromFloat(-1)))
		}
		rows[i].diff = diff
		balance = r.balance
	}
	return rows, nil
}

func main() {
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
			log.Fatal(err)
		}
		rows = append(rows, r...)
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
