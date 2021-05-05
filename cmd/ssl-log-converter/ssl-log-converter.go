package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
)

var extractGeometry = flag.Bool("extractGeometry", false, "Extract geometry messages into a new file")
var extractDetection = flag.Bool("extractDetection", false, "Extract detection messages into a new file")
var indentOutput = flag.Bool("indentOutput", false, "Indent the json-formatted output")

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

	for _, arg := range args {
		log.Println("Processing", arg)
		logReader, err := persistence.NewReader(arg)
		if err != nil {
			log.Fatalln(err)
		}

		f, err := os.Create(arg + ".txt")
		if err != nil {
			log.Fatalln("Could not open output file:", err)
		}
		channel := logReader.CreateChannel()

		for r := range channel {
			if r.MessageType.Id == persistence.MessageSslVision2014 {
				var visionMsg vision.SSL_WrapperPacket
				if err := proto.Unmarshal(r.Message, &visionMsg); err != nil {
					log.Println("Could not parse vision wrapper message:", err)
					continue
				}
				if *extractGeometry && visionMsg.Geometry != nil {
					writeMessage(f, visionMsg.Geometry)
				}
				if *extractDetection && visionMsg.Detection != nil {
					writeMessage(f, visionMsg.Detection)
				}
			}
		}

		if err := f.Close(); err != nil {
			log.Println("Could not close file:", err)
		}
	}
}

func writeMessage(f *os.File, v interface{}) {
	var b []byte
	var err error
	if *indentOutput {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		log.Println("Could not marshall detection:", err)
	} else {
		_, err = f.Write(b)
		check(err)
		_, err = f.WriteString("\n")
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Println("Unexpected error:", err)
	}
}
