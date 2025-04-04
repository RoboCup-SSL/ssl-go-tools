package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"log"
	"os"
	"strconv"
)

type RefereeProcessor struct {
	outFile         *os.File
	firstRefereeMsg *gc.Referee
	lastRefereeMsg  *gc.Referee

	FrameProcessor
}

func (p *RefereeProcessor) Init(logFile string) error {
	f, err := os.Create(logFile + "_referee.txt")
	if err != nil {
		return err
	}
	p.outFile = f

	return nil
}

func (p *RefereeProcessor) Close() error {

	if p.lastRefereeMsg != nil {
		if _, err := p.outFile.WriteString("\n\nLast message:\n"); err != nil {
			log.Println("Could not write to output file", err)
		}
		if b, err := prototext.Marshal(p.lastRefereeMsg); err != nil {
			log.Println("Could not marshal referee message: ", err)
		} else if _, err := p.outFile.WriteString(string(b)); err != nil {
			log.Println("Could not write referee message to output file: ", err)
		}
	}

	if p.outFile != nil {
		if err := p.outFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (p *RefereeProcessor) ProcessDetection(_ *persistence.Message, _ *vision.SSL_DetectionFrame) {
	// Not used
}

func (p *RefereeProcessor) ProcessReferee(_ *persistence.Message, frame *gc.Referee) {
	if p.lastRefereeMsg == nil {
		p.firstRefereeMsg = frame
		if _, err := p.outFile.WriteString("First message:\n"); err != nil {
			log.Println("Could not write to output file", err)
		}
		if b, err := prototext.Marshal(frame); err != nil {
			log.Println("Could not marshal referee message: ", err)
		} else if _, err := p.outFile.WriteString(string(b)); err != nil {
			log.Println("Could not write referee message to output file: ", err)
		}
	} else {
		if *frame.PacketTimestamp < *p.lastRefereeMsg.PacketTimestamp {
			log.Printf("Found smaller packet timestamp than last packet timestamp: \n%v\n%v",
				protojson.Format(p.lastRefereeMsg),
				protojson.Format(frame),
			)
		}
	}

	p.lastRefereeMsg = frame
}

func (p *RefereeProcessor) String() (res string) {
	if p.firstRefereeMsg == nil || p.lastRefereeMsg == nil {
		return
	}
	res += "First: " + p.firstRefereeMsg.Stage.String() + " " + strconv.Itoa(int(*p.firstRefereeMsg.StageTimeLeft/1000000)) + "s left\n"
	res += "Last: " + p.lastRefereeMsg.Stage.String() + " " + strconv.Itoa(int(*p.lastRefereeMsg.StageTimeLeft/1000000)) + "s left\n"
	res += "Duration: " + fmt.Sprintf("%.2f min\n", float64(*p.lastRefereeMsg.PacketTimestamp-*p.firstRefereeMsg.PacketTimestamp)/1e6/60)
	res += "Match type: " + p.lastRefereeMsg.GetMatchType().String()
	return
}
