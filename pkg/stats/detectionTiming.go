package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
)

const maxDt = 0.080

type DetectionTimingProcessor struct {
	lastDetection  *sslproto.SSL_DetectionFrame
	lastLogMessage *persistence.Message

	NumDetection uint64

	TCaptureDiffSum float64
	TSentDiffSum    float64
	TReceiveDiffSum float64

	NumCaptureDtOutlyer uint64
	NumSentDtOutlyer    uint64
	NumReceiveDtOutlyer uint64

	FrameProcessor
}

func (p *DetectionTimingProcessor) Init(string) error {
	return nil
}

func (p *DetectionTimingProcessor) Close() error {
	return nil
}

func (p *DetectionTimingProcessor) ProcessDetection(logMessage *persistence.Message, frame *sslproto.SSL_DetectionFrame) {
	if p.lastDetection != nil && p.lastLogMessage != nil {
		tCaptureDiff := *frame.TCapture - *p.lastDetection.TCapture
		tSentDiff := *frame.TSent - *p.lastDetection.TSent
		tReceiveDiff := float64(logMessage.Timestamp-p.lastLogMessage.Timestamp) / 1e9
		p.TCaptureDiffSum += tCaptureDiff
		p.TSentDiffSum += tSentDiff
		p.TReceiveDiffSum += tReceiveDiff

		if tCaptureDiff > maxDt {
			p.NumCaptureDtOutlyer++
		}
		if tSentDiff > maxDt {
			p.NumSentDtOutlyer++
		}
		if tReceiveDiff > maxDt {
			p.NumReceiveDtOutlyer++
		}
	}
	p.NumDetection++
	p.lastDetection = frame
	p.lastLogMessage = logMessage
}

func (p *DetectionTimingProcessor) ProcessReferee(*persistence.Message, *sslproto.Referee) {
}

func (p *DetectionTimingProcessor) String() (res string) {
	res = fmt.Sprintf("Detection frames: %d", p.NumDetection)
	res += fmt.Sprintf("\navg tCapture: %.4f", p.TCaptureDiffSum/float64(p.NumDetection))
	res += fmt.Sprintf("\navg tSent: %.4f", p.TSentDiffSum/float64(p.NumDetection))
	res += fmt.Sprintf("\navg tReceive: %.4f", p.TReceiveDiffSum/float64(p.NumDetection))
	res += fmt.Sprintf("\nNumber of frames with dt >%.4f -> capture: %d, sent: %d, receive: %d", maxDt, p.NumCaptureDtOutlyer, p.NumSentDtOutlyer, p.NumReceiveDtOutlyer)
	return
}
