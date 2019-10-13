package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/matchstats"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
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

	generator := matchstats.NewMatchStatsGenerator()
	for _, filename := range args {
		log.Println("Processing", filename)

		_, err := generator.Process(filename)
		if err != nil {
			log.Printf("%v: %v\n", filename, err)
		} else {
			log.Printf("Processed %v\n", filename)
		}

		log.Println("Processed", filename)
	}

	f, err := os.Create("out.json")
	if err != nil {
		log.Fatal("Could not create JSON output file", err)
	}

	jsonMarsh := jsonpb.Marshaler{EmitDefaults: true, Indent: "  "}
	err = jsonMarsh.Marshal(f, &generator.Collection)
	if err != nil {
		log.Println("Could not marshal match stats to json:", err)
	}

	f, err = os.Create("out.bin")
	if err != nil {
		log.Fatal("Could not create ProtoBuf output file", err)
	}

	bytes, err := proto.Marshal(&generator.Collection)
	if err != nil {
		log.Fatal("Could not marshal match stats to protobuf", err)
	}
	_, err = f.Write(bytes)
	if err != nil {
		log.Fatal("Could not write to protobuf output file", err)
	}
}

func usage() {
	_, err := fmt.Fprintln(os.Stderr, "Pass one or more log files that should be processed.")
	if err != nil {
		fmt.Println("Pass one or more log files that should be processed.")
	}
	flag.PrintDefaults()
}
