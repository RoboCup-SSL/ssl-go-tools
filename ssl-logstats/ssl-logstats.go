package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"log"
)

var useDetectionTimingExport = flag.Bool("detectionTimingExport", false, "Use this processor")
var useDetectionTiming = flag.Bool("detectionTiming", false, "Use this processor")

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatalln("Pass one or more log files")
	}

	for _, arg := range args {
		log.Println("Processing", arg)
		processLogFile(arg)
	}
}

func processLogFile(logFile string) {
	logReader, err := sslproto.NewLogReader(logFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer logReader.Close()

	channel := make(chan *sslproto.LogMessage, 100)
	go logReader.CreateLogMessageChannel(channel)

	var processors []FrameProcessor
	if useDetectionTimingExport != nil && *useDetectionTimingExport {
		processors = append(processors, new(DetectionTimingExportProcessor))
	}
	if useDetectionTiming != nil && *useDetectionTiming {
		processors = append(processors, new(DetectionTimingProcessor))
	}

	for _, p := range processors {
		err := p.Init(logFile)
		if err != nil {
			log.Println("Could not init processor:", err)
			return
		}
	}

	for r := range channel {
		if r.MessageType == sslproto.MESSAGE_SSL_VISION_2014 {
			visionMsg := r.ParseVisionWrapper()
			if visionMsg.Detection != nil {
				for _, p := range processors {
					p.ProcessDetection(r, visionMsg.Detection)
				}
			}
		}
	}

	for _, p := range processors {
		log.Println(p)
		p.Close()
	}
}
