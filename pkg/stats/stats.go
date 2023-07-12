package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
)

type FrameProcessor interface {
	ProcessDetection(*persistence.Message, *vision.SSL_DetectionFrame)
	ProcessReferee(*persistence.Message, *referee.Referee)
	Init(logFile string) error
	io.Closer
}

type Processor struct {
	UseAll                   bool
	UseDetectionTimingExport bool
	UseDetectionTiming       bool
	UseDetectionQuality      bool
	UseReferee               bool
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
	if allProcessors || p.UseReferee {
		processors = append(processors, new(RefereeProcessor))
	}

	for _, p := range processors {
		err := p.Init(logFile)
		if err != nil {
			log.Println("Could not init processor:", err)
			return
		}
	}

	numFrames := map[persistence.MessageId]int{}
	for r := range channel {
		numFrames[r.MessageType.Id]++
		if r.MessageType.Id == persistence.MessageSslVision2014 {

			var visionMsg vision.SSL_WrapperPacket
			if err := proto.Unmarshal(r.Message, &visionMsg); err != nil {
				log.Println("Could not parse vision wrapper message:", err)
				continue
			}
			if visionMsg.Detection != nil {
				for _, p := range processors {
					p.ProcessDetection(r, visionMsg.Detection)
				}
			}
		} else if r.MessageType.Id == persistence.MessageSslRefbox2013 {
			var refereeMsg referee.Referee
			if err := proto.Unmarshal(r.Message, &refereeMsg); err != nil {
				log.Println("Could not parse referee message: ", err)
				continue
			}
			for _, p := range processors {
				p.ProcessReferee(r, &refereeMsg)
			}
		}
	}

	log.Printf("Frames processed: %v", numFrames)

	for _, p := range processors {
		fmt.Println(p)
		fmt.Println()
		if err := p.Close(); err != nil {
			log.Println("Could not close processor: ", err)
		}
	}
}
