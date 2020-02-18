package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
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

func (p *DetectionTimingExportProcessor) ProcessDetection(logMessage *persistence.Message, frame *sslproto.SSL_DetectionFrame) {
	_, err := p.file.WriteString(fmt.Sprintf("%v,%v,%.30f,%.30f\n", logMessage.Timestamp, *frame.CameraId, *frame.TCapture, *frame.TSent))
	if err != nil {
		log.Println("Could not write timing: ", err)
	}
}

func (p *DetectionTimingExportProcessor) ProcessReferee(*persistence.Message, *sslproto.Referee) {
}

func (p *DetectionTimingExportProcessor) String() string {
	return ""
}
