package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"io"
	"log"
)

type FrameProcessor interface {
	ProcessDetection(*persistence.Message, *sslproto.SSL_DetectionFrame)
	ProcessReferee(*persistence.Message, *sslproto.Referee)
	Init(logFile string) error
	io.Closer
}

type Processor struct {
	UseAll                   bool
	UseDetectionTimingExport bool
	UseDetectionTiming       bool
	UseDetectionQuality      bool
	PrintQualityDataLosses   bool
}

func (p Processor) ProcessFile(logFile string) {
	logReader, err := persistence.NewReader(logFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := logReader.Close(); err != nil {
			log.Println("Could not close log reader: ", err)
		}
	}()

	channel := logReader.CreateChannel()

	allProcessors := p.UseAll

	var processors []FrameProcessor
	if allProcessors || p.UseDetectionTimingExport {
		processors = append(processors, new(DetectionTimingExportProcessor))
	}
	if allProcessors || p.UseDetectionTiming {
		processors = append(processors, new(DetectionTimingProcessor))
	}
	if allProcessors || p.UseDetectionQuality {
		proc := new(DetectionQualityProcessor)
		processors = append(processors, proc)
		proc.PrintDataLosses = p.PrintQualityDataLosses
	}

	for _, p := range processors {
		err := p.Init(logFile)
		if err != nil {
			log.Println("Could not init processor:", err)
			return
		}
	}

	for r := range channel {
		if r.MessageType.Id == persistence.MessageSslVision2014 {
			visionMsg, err := r.ParseVisionWrapper()
			if err != nil {
				log.Println("Could not parse vision wrapper message:", err)
				continue
			}
			if visionMsg.Detection != nil {
				for _, p := range processors {
					p.ProcessDetection(r, visionMsg.Detection)
				}
			}
		} else if r.MessageType.Id == persistence.MessageSslRefbox2013 {
			refereeMsg, err := r.ParseReferee()
			if err != nil {
				log.Println("Could not parse referee massage: ", err)
				continue
			}
			for _, p := range processors {
				p.ProcessReferee(r, refereeMsg)
			}
		}
	}

	for _, p := range processors {
		fmt.Println(p)
		if err := p.Close(); err != nil {
			log.Println("Could not close processor: ", err)
		}
	}
}
