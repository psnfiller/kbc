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

func classify(in string) string {
	switch {
	case in == "Non-Euro Point Of Sales Fee":
		return "Fee"

	case in == "SDD KBC Bank Ireland Public Limited":
		return "House"
	case in == "SDD KBC Bank Ireland Public Limite":
		return "House"
	case strings.Contains(in, " PROPERTY TAX"):
		return "House"
	case strings.Contains(in, "POS THORNTONS RECYCLING"):
		return "House"
	case strings.Contains(in, "SDD ELECTRIC IRELAND"):
		return "House"
	case strings.Contains(in, "Flogas Natural Gas"):
		return "House"
	case strings.Contains(in, "FLOGAS NATURAL GAS"):
		return "House"
	case strings.Contains(in, "EIRCOM"):
		return "House"
	case strings.Contains(in, "EIR"):
		return "House"
	case strings.Contains(in, "IRISH LIFE"):
		return "House"
	case strings.Contains(in, "Irish Life"):
		return "House"
	case strings.Contains(in, "SCT KBC MORTGAGE"):
		return "House"
	case strings.Contains(in, "Irish Water"):
		return "House"
	case strings.Contains(in, "residents association"):
		return "House"

	case strings.Contains(in, "BROWN THOMAS"):
		return "House stuff"
	case strings.Contains(in, "EZLIVING"):
		return "House stuff"
	case strings.Contains(in, "POWERCITY.IE"):
		return "House stuff"
	case strings.Contains(in, "tiling"):
		return "House stuff"
	case strings.Contains(in, "HOUSE OF FRASER"):
		return "House stuff"
	case strings.Contains(in, "ONCEOFFPAYMENT plumbing"):
		return "House stuff"
	case strings.Contains(in, "IKEA"):
		return "House stuff"
	case strings.Contains(in, "WOODIES"):
		return "house stuff"
	case strings.Contains(in, "HOWARDS STORAGE WORL"):
		return "house stuff"
	case strings.Contains(in, "OS ITALIAN TILE"):
		return "house stuff"

	case strings.Contains(in, " WWW.THEWITCHERY.CO"):
		return "presents"
	case strings.Contains(in, " FORTNUMANDMASON.CO"):
		return "presents"

	case strings.Contains(in, "Peter Nuttall to rbs"):
		return "rbs"
	case strings.Contains(in, "Peter Nuttall Dashboard"):
		return "rbs"

	case strings.Contains(in, "CHADWICKS"):
		return "garden"
	case strings.Contains(in, "ARBORETUM KILQUADE "):
		return "garden"
	case strings.Contains(in, "ERHS PLANTS"):
		return "garden"
	case strings.Contains(in, "DAVID AUSTIN ROSE "):
		return "garden"
	case strings.Contains(in, "WWW.GREENGARDENER.CO"):
		return "garden"

	case strings.Contains(in, "BEERS CHEERS"):
		return "holidays"
	case strings.Contains(in, "Johan Nystro"):
		return "holidays"
	case strings.Contains(in, "DJURGARDSBRON"):
		return "holidays"
	case strings.Contains(in, "SIOP Y LLAN "):
		return "holidays"
	case strings.Contains(in, "JUNIBACKEN-ENTRE"):
		return "holidays"
	case strings.Contains(in, "ICA SUPERMARKET HOGA"):
		return "holidays"
	case strings.Contains(in, "REWE"):
		return "holidays"
	case strings.Contains(in, "NOBELMUSEET"):
		return "holidays"
	case strings.Contains(in, "Vasamuseets Restaura"):
		return "holidays"
	case strings.Contains(in, "TALYLLYN RAILWAY"):
		return "holidays"
	case strings.Contains(in, "VAPIANO Mainz Rheint"):
		return "holidays"
	case strings.Contains(in, "Trainline.com"):
		return "holidays"
	case strings.Contains(in, "ART OFFICE STOCKHO"):
		return "holidays"
	case strings.Contains(in, "GRAY LINE TOURS"):
		return "holidays"
	case strings.Contains(in, "TALYLLYN RAILWAY"):
		return "holidays"
	case strings.Contains(in, "Trainline.com"):
		return "holidays"
	case strings.Contains(in, "VASAMUSEETS"):
		return "holidays"
	case strings.Contains(in, "INTREPID SEA AIR AND"):
		return "holidays"
	case strings.Contains(in, "IRISH FERRIE"):
		return "holidays"
	case strings.Contains(in, "DAA CARPARK SERVIC"):
		return "holidays"
	case strings.Contains(in, "DUBLIN AIRPORT BUS"):
		return "holidays"
	case strings.Contains(in, "VILLAGGIO "):
		return "holidays"
	case strings.Contains(in, " PIZZERIA ALBONA"):
		return "holidays"
	case strings.Contains(in, "Oberwesel"):
		return "holidays"
	case strings.Contains(in, "COMPUTER HISTORY"):
		return "holidays"
	case strings.Contains(in, "Stena Lin"):
		return "holidays"
	case strings.Contains(in, "ROYAL OAK HOTEL"):
		return "holidays"
	case strings.Contains(in, "TRAFALGAR HOUSE"):
		return "holidays"
	case strings.Contains(in, "HOTEL WEINBERG SCHLO"):
		return "holidays"
	case strings.Contains(in, "DUBAIRPORT"):
		return "holidays"
	case strings.Contains(in, "BABYLOFT"):
		return "holidays"
	case strings.Contains(in, "SCHWEIZERKONDITORI"):
		return "holidays"
	case strings.Contains(in, "JOSEPH BRANNIGAN"):
		return "holidays"
	case strings.Contains(in, "LYFT RIDE"):
		return "holidays"
	case strings.Contains(in, "DEBENHAMS"):
		return "holidays"
	case strings.Contains(in, "AMTRAK"):
		return "holidays"
	case strings.Contains(in, "MEINHOLF"):
		return "holidays"
	case strings.Contains(in, "RYANAIR"):
		return "holidays"
	case strings.Contains(in, "HOTELL ZINKENSDAMM"):
		return "holidays"
	case strings.Contains(in, "AIRBNB"):
		return "holidays"
	case strings.Contains(in, "LUFTHANSA"):
		return "holidays"
	case strings.Contains(in, "TRIPADV"):
		return "holidays"
	case strings.Contains(in, "AERLING"):
		return "holidays"
	case strings.Contains(in, "SAS"):
		return "holidays"
	case strings.Contains(in, "AVIS RENT-A-CAR"):
		return "holidays"
	case strings.Contains(in, "ARLANDA"):
		return "holidays"
	case strings.Contains(in, "IRISH FERRIES LIMITE"):
		return "holidays"
	case strings.Contains(in, "NOVOTEL BRUSSELS OFF"):
		return "holidays"
	case strings.Contains(in, "STABLE COURT APARTME"):
		return "holidays"
	case strings.Contains(in, "RENTALCARS.COM"):
		return "holidays"
	case strings.Contains(in, "Stena Line Ltd"):
		return "holidays"
	case strings.Contains(in, "PARK INN BELFAST"):
		return "holidays"
	case strings.Contains(in, "CSH ICE AIRSIDE DEPARTUR"):
		return "holidays"
	case strings.Contains(in, "POS SNOW ROCK"):
		return "holidays"

	case strings.Contains(in, "TOMTHUMBBAB"):
		return "toys"
	case strings.Contains(in, "WWW.MYRIADONLINE.CO"):
		return "toys"
	case strings.Contains(in, "SMYTHS TOYS"):
		return "toys"
	case strings.Contains(in, "KIDDING AROUND"):
		return "toys"
	case strings.Contains(in, "BABIES R US"):
		return "toys"
	case strings.Contains(in, "ONE HUNDRED TOYS"):
		return "toys"
	case strings.Contains(in, "SP CONSCIOUS CRAFT"):
		return "toys"
	case strings.Contains(in, " WWW.BABIPUR.CO.UK"):
		return "toys"
	case strings.Contains(in, "MILLETS"):
		return "toys"

	case strings.Contains(in, "CYCLE PLUS"):
		return "bike"

	case strings.Contains(in, "MOVEMBER CHARITY"):
		return "charity"

	case strings.Contains(in, "DONNYBROOK FAIR"):
		return "Food"
	case strings.Contains(in, "CO-OP"):
		return "Food"
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
	case strings.Contains(in, "fortnumandmason.com"):
		return "Food"
	case strings.Contains(in, "CAVISTONS"):
		return "Food"
	case strings.Contains(in, "SUPERVALU"):
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
	case strings.Contains(in, "RASAM"):
		return "meals out"
	case strings.Contains(in, "BAREBURGER"):
		return "meals out"
	case strings.Contains(in, "BOX BURGER"):
		return "meals out"
	case strings.Contains(in, "BENITOS RESTAURANT"):
		return "meals out"
	case strings.Contains(in, "KATHMANDU"):
		return "meals out"
	case strings.Contains(in, "THE HUNGRY MONK REST"):
		return "meals out"
	case strings.Contains(in, "GLENMALURE LODGE"):
		return "meals out"

	case strings.Contains(in, "THE PORTER HOUS"):
		return "booze"
	case strings.Contains(in, "O BRIENS WINES"):
		return "booze"
	case strings.Contains(in, "WWW.BEERGONZO.CO.UK"):
		return "booze"
	case strings.Contains(in, "64 WINE"):
		return "booze"
	case strings.Contains(in, "THE MARTELLO HOTEL"):
		return "booze"
	case strings.Contains(in, "BLACKROCK CELLAR"):
		return "booze"
	case strings.Contains(in, "IDLEWILDE"):
		return "booze"

	case strings.Contains(in, "POS GREAT OUTDOORS"):
		return "clothes"
	case strings.Contains(in, "HOTMILK LINGERIE"):
		return "clothes"
	case strings.Contains(in, "MARKS SPENCER"):
		return "clothes"
	case strings.Contains(in, "EX OFFICIO"):
		return "clothes"
	case strings.Contains(in, "SERAPHINE"):
		return "clothes"
	case strings.Contains(in, "WWW.FATFACE.COM"):
		return "clothes"
	case strings.Contains(in, "Marks and Spencer"):
		return "clothes"
	case strings.Contains(in, "NEXT"):
		return "clothes"

	case strings.Contains(in, "Footprints Montessori"):
		return "childcare"
	case strings.Contains(in, "domiso"):
		return "childcare"
	case strings.Contains(in, "DOMISOMUSIC"):
		return "childcare"
	case strings.Contains(in, "preschool"):
		return "childcare"

	case strings.Contains(in, "MOTHERCARE"):
		return "baby clothes"
	case strings.Contains(in, "JOJOMAMANBEBE.CO.U"):
		return "baby clothes"
	case strings.Contains(in, "JOJO MAMAN BEBE"):
		return "baby clothes"
	case strings.Contains(in, "HE CLARKS SHOP"):
		return "baby clothes"

	case strings.Contains(in, "CHARLESLAND SPORT"):
		return "wife "

	case strings.Contains(in, "Kia Ora"):
		return "day out"
	case strings.Contains(in, "IMAGINOSITY"):
		return "day out"
	case strings.Contains(in, "NCH.IE"):
		return "day out"
	case strings.Contains(in, "DUBLIN ZOO"):
		return "day out"
	case strings.Contains(in, "TICKET MASTER"):
		return "day out"
	case strings.Contains(in, "GLENROE"):
		return "day out"
	case strings.Contains(in, "KILLRUDDERY"):
		return "day out"
	case strings.Contains(in, "Zoom Adventure"):
		return "day out"
	case strings.Contains(in, "POWERSCOURT"):
		return "day out"
	case strings.Contains(in, "THE WETLAND CENTRE"):
		return "day out"
	case strings.Contains(in, "GREENS BERRY FARM"):
		return "day out"
	case strings.Contains(in, "ROYAL ZOOLOGICAL SOC"):
		return "day out"

	case strings.Contains(in, " Google Ireland Limited"):
		return "Pay"
	case strings.Contains(in, "Transfer MORGAN STANLEY SMITH"):
		return "Pay"
	case strings.Contains(in, "Transfer MORGAN STNLEY SMITH"):
		return "Pay"
	case strings.Contains(in, "Transfer MORGAN STANLEY SMIT"):
		return "Pay"
	case strings.Contains(in, "GOOGLE"):
		return "Pay"
	case strings.Contains(in, "BIRDWATCH IRELAND BWI"):
		return "Pay"

	case strings.Contains(in, "KBC Online Dashboard Transfer"):
		return "transfer in"
	case strings.Contains(in, "SCT PETER NUTTALL RBS"):
		return "transfer in"
	case strings.Contains(in, "KBC Mobile from smart online"):
		return "transfer in"
	case strings.Contains(in, "KBC Online fornewregsave"):
		return "transfer in"

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
	case strings.Contains(in, "KBC Mobile to smart online"):
		return "Savings"
	case strings.Contains(in, "Online SCT new regular saver"):
		return "Savings"

	case strings.Contains(in, "MOTHERJONES"):
		return "News"

	case strings.Contains(in, "SEEEDSTUDIO"):
		return "Computers"
	case strings.Contains(in, "UPUTRONICS"):
		return "Computers"
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
	case strings.Contains(in, "TARSNAP.COM"):
		return "Computers"

	case strings.Contains(in, "APPLE STORE"):
		return "Phone"

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

	case strings.HasPrefix(in, "POS Montag Steins Cloc"):
		return "CLOCK"

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

	case strings.Contains(in, "PHARMACY"):
		return "medical"
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
	case strings.Contains(in, "GREYSTONES EYE CENTR"):
		return "medical"
	case strings.Contains(in, "OUR LADYS HOSP CRUM"):
		return "medical"
	case strings.Contains(in, "THE ULTRASOUND SUIT"):
		return "medical"
	case strings.Contains(in, "VHI Claims Vhi"):
		return "medical"

	case strings.Contains(in, "OREILLY"):
		return "books"
	case strings.Contains(in, "STRAND BOOK STORE"):
		return "books"
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
	case strings.Contains(in, "HODGES FIGGIS"):
		return "books"
	case strings.Contains(in, "BOOKS INC"):
		return "books"
	case strings.Contains(in, "THE GUTTER BOOKSHOP"):
		return "books"
	case strings.Contains(in, "REFORMATION HERITAGE"):
		return "books"
	case strings.Contains(in, "MEDIA GRATIAE"):
		return "books"
	case strings.Contains(in, "CHRISTIAN FOCUS"):
		return "books"

	case strings.Contains(in, "INTERFLORA"):
		return "flowers"
	case strings.Contains(in, "COTTAGE FLOWER"):
		return "flowers"

	case strings.Contains(in, "AN POST-OFFICES"):
		return "post"

	case strings.Contains(in, "HALFORDS"):
		return "car"
	case strings.Contains(in, "APPLEGREEN MOTORWAY"):
		return "car"
	case strings.Contains(in, "AVIVA DIRECT"):
		return "car"
	case strings.Contains(in, "WWW.AVIVA.IE"):
		return "car"
	case strings.Contains(in, "123 MONEY"):
		return "car"
	case strings.Contains(in, "HILLS OF GREYSTONES"):
		return "car"
	case strings.Contains(in, "ONLINE MOTOR TAX"):
		return "car"
	case strings.Contains(in, "Greystones Exhausts"):
		return "car"
	case strings.Contains(in, "CAR VALETING"):
		return "car"
	case strings.Contains(in, "TOPAZ"):
		return "car"
	case strings.Contains(in, "CENTRA"):
		return "car"
	case strings.Contains(in, "APPLEGREEN"):
		return "car"
	case strings.Contains(in, "WWW NCTS IE"):
		return "car"
	case strings.Contains(in, "ESSO"):
		return "car"
	case strings.Contains(in, "M11 WICKLOW SERVICE"):
		return "car"

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
			return rows, fmt.Errorf("failed to do diff %s %s %s %s", r, diff, r.change, invert)
		}
		rows[i].diff = diff
		balance = r.balance
	}
	return rows, nil
}

func main() {
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
