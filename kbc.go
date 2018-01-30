package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`\s+(\d\d [A-Z][a-z]{2} 201\d)\s+(.*)\s\s`)
	f := "/Users/psn/Downloads/wat/Current Account Statement - 01 Oct 2017.txt"
	fd, err := os.Open(f)
	defer fd.Close()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		l := re.FindStringSubmatch(line)
		if l == nil {
			continue
		}
		fmt.Printf("%s %s\n", l[1], l[2])

	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
