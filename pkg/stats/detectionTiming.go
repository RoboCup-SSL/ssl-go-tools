package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
)

const maxDt = 0.080

type DetectionTimingProcessor struct {
	lastLogMessage *persistence.Message

	cameraTimings map[uint32]*CameraTiming

	NumDetection        uint64
	TReceiveDiffSum     float64
	NumReceiveDtOutlyer uint64

	FrameProcessor
}

type CameraTiming struct {
	lastDetection *vision.SSL_DetectionFrame

	NumDetection uint64

	TCaptureDiffSum float64
	TSentDiffSum    float64

	NumCaptureDtOutlyer uint64
	NumSentDtOutlyer    uint64
}

func (p *DetectionTimingProcessor) Init(string) error {
	p.cameraTimings = map[uint32]*CameraTiming{}
	return nil
}

func (p *DetectionTimingProcessor) Close() error {
	return nil
}

func (p *DetectionTimingProcessor) ProcessDetection(logMessage *persistence.Message, frame *vision.SSL_DetectionFrame) {
	if p.lastLogMessage != nil {
		tReceiveDiff := float64(logMessage.Timestamp-p.lastLogMessage.Timestamp) / 1e9
		p.TReceiveDiffSum += tReceiveDiff

		if tReceiveDiff > maxDt {
			p.NumReceiveDtOutlyer++
		}
	}

	cameraTiming := p.cameraTimings[*frame.CameraId]
	if cameraTiming == nil {
		cameraTiming = new(CameraTiming)
		p.cameraTimings[*frame.CameraId] = cameraTiming
	}
	cameraTiming.Process(frame)

	p.NumDetection++
	p.lastLogMessage = logMessage
}

func (p *CameraTiming) Process(frame *vision.SSL_DetectionFrame) {
	if p.lastDetection != nil {
		tCaptureDiff := *frame.TCapture - *p.lastDetection.TCapture
		tSentDiff := *frame.TSent - *p.lastDetection.TSent
		p.TCaptureDiffSum += tCaptureDiff
		p.TSentDiffSum += tSentDiff

		if tCaptureDiff > maxDt {
			p.NumCaptureDtOutlyer++
		}
		if tSentDiff > maxDt {
			p.NumSentDtOutlyer++
		}
	}
	p.NumDetection++
	p.lastDetection = frame
}

func (p *DetectionTimingProcessor) ProcessReferee(*persistence.Message, *referee.Referee) {
}

func (p *DetectionTimingProcessor) String() (res string) {
	res = fmt.Sprintf("Overall frames: %d", p.NumDetection)
	res += fmt.Sprintf("\navg tReceive: %.4f", p.TReceiveDiffSum/float64(p.NumDetection))
	res += fmt.Sprintf("\nNumber of frames with dt >%.4f -> receive: %d", maxDt, p.NumReceiveDtOutlyer)
	for cameraId, cameraTiming := range p.cameraTimings {
		res += fmt.Sprintf("\nCamera %v", cameraId)
		res += fmt.Sprintf("\nDetection frames: %d", cameraTiming.NumDetection)
		res += fmt.Sprintf("\navg tCapture: %.4f", cameraTiming.TCaptureDiffSum/float64(cameraTiming.NumDetection))
		res += fmt.Sprintf("\navg tSent: %.4f", cameraTiming.TSentDiffSum/float64(cameraTiming.NumDetection))
		res += fmt.Sprintf("\nNumber of frames with dt >%.4f -> capture: %d, sent: %d", maxDt, cameraTiming.NumCaptureDtOutlyer, cameraTiming.NumSentDtOutlyer)
	}
	return
}
