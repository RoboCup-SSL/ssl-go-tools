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
var addLogFileTimestamp = flag.Bool("addLogFileTimestamp", false, "Add the timestamp of the log file to the json output")

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
					check(writeMessage(f, r.Timestamp, visionMsg.Geometry))
				}
				if *extractDetection && visionMsg.Detection != nil {
					check(writeMessage(f, r.Timestamp, visionMsg.Detection))
				}
			}
		}

		if err := f.Close(); err != nil {
			log.Println("Could not close file:", err)
		}
	}
}

func writeMessage(f *os.File, timestamp int64, v interface{}) error {
	var result map[string]interface{}

	if b, err := json.Marshal(v); err != nil {
		return err
	} else {
		if err := json.Unmarshal(b, &result); err != nil {
			return err
		}
		if *addLogFileTimestamp {
			result["timestamp"] = timestamp
		}
	}

	var data []byte
	if *indentOutput {
		if b, err := json.MarshalIndent(result, "", "  "); err != nil {
			return err
		} else {
			data = b
		}
	} else {
		if b, err := json.Marshal(result); err != nil {
			return err
		} else {
			data = b
		}
	}

	if _, err := f.Write(data); err != nil {
		return err
	}
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func check(err error) {
	if err != nil {
		log.Println("Unexpected error:", err)
	}
}
