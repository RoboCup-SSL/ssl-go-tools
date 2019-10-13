package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/matchstats"
	"github.com/golang/protobuf/jsonpb"
	"log"
	"os"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	for _, filename := range args {
		log.Println("Processing", filename)

		generator := matchstats.NewMatchStatsGenerator()
		matchStats, err := generator.Process(filename)
		if err != nil {
			log.Printf("%v: %v\n", filename, err)
		} else {
			log.Printf("Processed %v\n", filename)
			marshaler := jsonpb.Marshaler{EmitDefaults: true, Indent: "  "}
			b, err := marshaler.MarshalToString(matchStats)
			if err != nil {
				log.Println("Could not marshal match stats to json:", err)
			}
			fmt.Println(string(b))
		}

		log.Println("Processed", filename)
	}
}

func usage() {
	_, err := fmt.Fprintln(os.Stderr, "Pass one or more log files that should be processed.")
	if err != nil {
		fmt.Println("Pass one or more log files that should be processed.")
	}
	flag.PrintDefaults()
}
