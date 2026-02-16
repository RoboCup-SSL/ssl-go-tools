package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"log"
)

var verify = flag.Bool("verify", false, "Verify the index")
var writeIndex = flag.Bool("index", true, "Create an index of the log file")
var writeGcIndex = flag.Bool("gc", false, "Create a gc index based on GC messages")
var writeMessageTypes = flag.Bool("messageTypes", false, "Create per-message-type indices")

func main() {
	flag.Usage = func() {
		fmt.Println("Pass one or more log files in.")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	for _, logfile := range args {
		processLogfile(logfile)
	}
}

func processLogfile(logfile string) {
	if *writeIndex {
		if err := index.WriteIndex(logfile); err != nil {
			log.Println("Could not index log file:", logfile, err)
		}
	}

	if *verify {
		verifyIndex(logfile)
	}

	if *writeGcIndex {
		if err := index.WriteGcIndex(logfile); err != nil {
			log.Println("Could not create gc index:", logfile, err)
		}
	}

	if *writeMessageTypes {
		if err := index.WriteMessageTypeIndices(logfile); err != nil {
			log.Println("Could not create message type indices:", logfile, err)
		}
	}
}

func verifyIndex(logfile string) {
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
