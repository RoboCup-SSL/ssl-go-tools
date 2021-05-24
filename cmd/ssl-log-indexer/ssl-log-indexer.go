package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"log"
)

var verify = flag.Bool("verify", false, "Verify the index")

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

		if *verify {
			reader, _ := persistence.NewReader(logfile)
			offsets, err := reader.ReadIndex()
			if err != nil {
				panic(err)
			}
			log.Printf("Index size: %d", len(offsets))

			n := 1
			msg, err := reader.ReadMessageAt(offsets[n])
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Message %d: %v", n, *msg)
			}
		}
	}
}
