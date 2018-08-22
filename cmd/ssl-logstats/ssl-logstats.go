package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/stats"
	"log"
)

var useDetectionTimingExport = flag.Bool("detectionTimingExport", false, "Use this processor")
var useDetectionTiming = flag.Bool("detectionTiming", false, "Use this processor")
var useAll = flag.Bool("all", false, "Use all processors")

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatalln("Pass one or more log files")
	}

	p := stats.Processor{}
	p.UseAll = *useAll
	p.UseDetectionTiming = *useDetectionTiming
	p.UseDetectionTimingExport = *useDetectionTimingExport

	for _, arg := range args {
		log.Println("Processing", arg)
		p.ProcessFile(arg)
	}
}
