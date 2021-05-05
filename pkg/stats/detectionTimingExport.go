package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"log"
	"os"
)

type DetectionTimingExportProcessor struct {
	file *os.File

	FrameProcessor
}

func (p *DetectionTimingExportProcessor) Init(logFile string) error {
	f, err := os.Create(logFile + ".csv")
	if err != nil {
		return err
	}
	p.file = f
	return nil
}

func (p *DetectionTimingExportProcessor) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

func (p *DetectionTimingExportProcessor) ProcessDetection(logMessage *persistence.Message, frame *vision.SSL_DetectionFrame) {
	_, err := p.file.WriteString(fmt.Sprintf("%v,%v,%.30f,%.30f\n", logMessage.Timestamp, *frame.CameraId, *frame.TCapture, *frame.TSent))
	if err != nil {
		log.Println("Could not write timing: ", err)
	}
}

func (p *DetectionTimingExportProcessor) ProcessReferee(*persistence.Message, *referee.Referee) {
}

func (p *DetectionTimingExportProcessor) String() string {
	return ""
}
