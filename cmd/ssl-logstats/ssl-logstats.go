package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/stats"
	"log"
)

var useDetectionTimingExport = flag.Bool("detectionTimingExport", false, "Use this processor")
var useDetectionTiming = flag.Bool("detectionTiming", false, "Use this processor")
var useDetectionQuality = flag.Bool("detectionQuality", false, "Use this processor")
var useReferee = flag.Bool("referee", false, "Use this processor")
var printQualityDataLosses = flag.Bool("printQualityDataLosses", false, "Print data losses over threshold from quality detector")
var useAll = flag.Bool("all", false, "Use all processors")

func main() {
	flag.Usage = func() {
		fmt.Println("Pass one or more log files and specify one or more processors with following flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatalln("Pass one or more log files")
	}

	p := stats.Processor{}
	p.UseAll = *useAll
	p.UseDetectionTiming = *useDetectionTiming
	p.UseDetectionTimingExport = *useDetectionTimingExport
	p.UseDetectionQuality = *useDetectionQuality
	p.UseReferee = *useReferee
	p.PrintQualityDataLosses = *printQualityDataLosses

	for _, arg := range args {
		log.Println("Processing", arg)
		p.ProcessFile(arg)
	}
}
