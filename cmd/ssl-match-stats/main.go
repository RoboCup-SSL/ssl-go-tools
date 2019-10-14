package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/matchstats"
	"log"
	"os"
)

func main() {
	flag.Usage = usage

	fGenerate := flag.Bool("generate", false, "Generate statistics based on passed in log files")
	fExportCsv := flag.Bool("exportCsv", false, "Export data from a generated out.bin file to CSV")

	flag.Parse()

	if *fGenerate {
		generate()
	} else if *fExportCsv {
		exportCsv()
	} else {
		usage()
	}
}

func generate() {
	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	a := matchstats.NewAggregator()
	for _, filename := range args {
		log.Println("Processing", filename)

		err := a.Process(filename)
		if err != nil {
			log.Printf("%v: %v\n", filename, err)
		} else {
			log.Printf("Processed %v\n", filename)
		}
	}

	if err := a.WriteBin("out.bin"); err != nil {
		log.Println("Could not write binary file", err)
	}

	if err := a.WriteJson("out.json"); err != nil {
		log.Println("Could not write JSON file", err)
	}
}

func exportCsv() {

	a := matchstats.NewAggregator()

	if err := a.ReadBin("out.bin"); err != nil {
		log.Fatal(err)
	}

	if err := matchstats.WriteGamePhaseDurations(&a.Collection, "game-phase-durations.csv"); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	_, err := fmt.Fprintln(os.Stderr, "Pass one or more log files that should be processed.")
	if err != nil {
		fmt.Println("Pass one or more log files that should be processed.")
	}
	flag.PrintDefaults()
}
