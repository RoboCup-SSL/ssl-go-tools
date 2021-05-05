package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"log"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Pass one or more log files in.")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	for _, logfile := range args {
		if err := index.WriteIndex(logfile); err != nil {
			log.Println("Could not index log file:", logfile, err)
		}
	}
}
