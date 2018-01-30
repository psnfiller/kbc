package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`\s+\d\d [A-Z][a-z]{2} 2017`)
	f := "/Users/psn/Downloads/wat/Current Account Statement - 01 Oct 2017.txt"
	fd, err := os.Open(f)
	defer fd.Close()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if re.FindString(line) != "" {
			fmt.Println(line)
		}

	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
