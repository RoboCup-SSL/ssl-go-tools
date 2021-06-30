package stats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"strconv"
)

type RefereeProcessor struct {
	outFile         *os.File
	firstRefereeMsg *referee.Referee
	lastRefereeMsg  *referee.Referee

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
		if err := proto.MarshalText(p.outFile, p.lastRefereeMsg); err != nil {
			log.Println("Could not write referee message to output file", err)
		}
	}

	if p.outFile != nil {
		if err := p.outFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (p *RefereeProcessor) ProcessDetection(_ *persistence.Message, frame *vision.SSL_DetectionFrame) {

}

func (p *RefereeProcessor) ProcessReferee(_ *persistence.Message, frame *referee.Referee) {
	if p.lastRefereeMsg == nil {
		p.firstRefereeMsg = frame
		if _, err := p.outFile.WriteString("First message:\n"); err != nil {
			log.Println("Could not write to output file", err)
		}
		if err := proto.MarshalText(p.outFile, frame); err != nil {
			log.Println("Could not write referee message to output file", err)
		}
	}

	p.lastRefereeMsg = frame
}

func (p *RefereeProcessor) String() (res string) {
	if p.firstRefereeMsg != nil {
		res += "First: " + p.firstRefereeMsg.Stage.String() + " " + strconv.Itoa(int(*p.firstRefereeMsg.StageTimeLeft)) + "\n"
	}
	if p.lastRefereeMsg != nil {
		res += "Last: " + p.lastRefereeMsg.Stage.String() + " " + strconv.Itoa(int(*p.lastRefereeMsg.StageTimeLeft)) + "\n"
	}
	return
}
