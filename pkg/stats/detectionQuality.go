package stats

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"log"
	"os"
	"sort"
	"time"
)

const maxRobotDt = time.Millisecond * 1000
const maxRobotLoss = time.Millisecond * 200
const minRobotAge = time.Second * 10

const maxBallDt = time.Millisecond * 1000
const maxBallLoss = time.Millisecond * 100
const minBallAge = time.Second * 5

type DetectionQualityProcessor struct {
	active          bool
	cameraData      map[uint32]*Camera
	robotDataLosses []*DataLoss
	ballDataLosses  []*DataLoss
	robotsFile      *os.File
	PrintDataLosses bool

	FrameProcessor
}

type Camera struct {
	id     uint32
	robots map[TeamColor]map[uint32]*Robot
	ball   *Ball
}

type TeamColor int

const (
	TeamYellow TeamColor = 1
	TeamBlue             = 2
)

type Robot struct {
	id                 uint32
	lastFrameId        uint32
	firstDetectionTime time.Time
	lastDetectionTime  time.Time
}

type Ball struct {
	lastFrameId        uint32
	firstDetectionTime time.Time
	lastDetectionTime  time.Time
}

type DataLoss struct {
	Time      time.Time
	Duration  time.Duration
	NumFrames uint32
	ObjectAge time.Duration
	RobotId   uint32
	TeamColor TeamColor
}

func (p *DetectionQualityProcessor) Init(logFile string) error {
	p.cameraData = map[uint32]*Camera{}

	f, err := os.Create(logFile + "_robots.csv")
	if err != nil {
		return err
	}
	p.robotsFile = f

	return nil
}

func (p *DetectionQualityProcessor) Close() error {

	for _, dataLoss := range p.robotDataLosses {
		_, err := p.robotsFile.WriteString(fmt.Sprintf("%v,%v,%v,%v,%v,%v\n", dataLoss.Time.UnixNano(), dataLoss.TeamColor, dataLoss.RobotId, dataLoss.ObjectAge.Nanoseconds(), dataLoss.NumFrames, dataLoss.Duration.Nanoseconds()))
		if err != nil {
			log.Println("Could not write timing: ", err)
			break
		}
	}

	if p.robotsFile != nil {
		return p.robotsFile.Close()
	}

	return nil
}

func (p *DetectionQualityProcessor) ProcessDetection(_ *persistence.Message, frame *sslproto.SSL_DetectionFrame) {
	if !p.active {
		return
	}

	camera := p.cameraData[*frame.CameraId]
	if camera == nil {
		camera = new(Camera)
		camera.id = *frame.CameraId
		camera.robots = map[TeamColor]map[uint32]*Robot{}
		camera.robots[TeamYellow] = map[uint32]*Robot{}
		camera.robots[TeamBlue] = map[uint32]*Robot{}
		p.cameraData[*frame.CameraId] = camera
	}

	if dataLoss := camera.processRobots(frame, frame.RobotsYellow, TeamYellow); dataLoss != nil {
		p.robotDataLosses = append(p.robotDataLosses, dataLoss)
	}
	if dataLoss := camera.processRobots(frame, frame.RobotsBlue, TeamBlue); dataLoss != nil {
		p.robotDataLosses = append(p.robotDataLosses, dataLoss)
	}
	if dataLoss := camera.processBalls(frame); dataLoss != nil {
		p.ballDataLosses = append(p.ballDataLosses, dataLoss)
	}
}

func (p *Camera) processRobots(frame *sslproto.SSL_DetectionFrame, robots []*sslproto.SSL_DetectionRobot, teamColor TeamColor) (dataLoss *DataLoss) {
	dataLoss = nil
	for _, detectionRobot := range robots {
		robot := p.robots[teamColor][*detectionRobot.RobotId]
		if robot == nil {
			robot = new(Robot)
			p.robots[teamColor][*detectionRobot.RobotId] = robot
			robot.id = *detectionRobot.RobotId
		}
		tSent := toTime(*frame.TSent)
		dt := tSent.Sub(robot.lastDetectionTime)
		frameDiff := *frame.FrameNumber - robot.lastFrameId
		if dt > maxRobotDt {
			robot.firstDetectionTime = tSent
		} else if frameDiff > 1 {
			dataLoss = &DataLoss{
				Duration:  dt,
				NumFrames: frameDiff,
				Time:      tSent,
				ObjectAge: tSent.Sub(robot.firstDetectionTime),
				RobotId:   robot.id,
				TeamColor: teamColor,
			}
		}
		robot.lastDetectionTime = tSent
		robot.lastFrameId = *frame.FrameNumber
	}
	return
}

func (p *Camera) processBalls(frame *sslproto.SSL_DetectionFrame) (dataLoss *DataLoss) {
	dataLoss = nil
	if len(frame.Balls) == 0 {
		return
	}
	if p.ball == nil {
		p.ball = new(Ball)
	}
	tSent := toTime(*frame.TSent)
	dt := tSent.Sub(p.ball.lastDetectionTime)
	frameDiff := *frame.FrameNumber - p.ball.lastFrameId
	if dt > maxBallDt {
		p.ball.firstDetectionTime = tSent
	} else if frameDiff > 1 {
		dataLoss = &DataLoss{
			Duration:  dt,
			NumFrames: frameDiff,
			Time:      tSent,
			ObjectAge: tSent.Sub(p.ball.firstDetectionTime),
		}
	}
	p.ball.lastDetectionTime = tSent
	p.ball.lastFrameId = *frame.FrameNumber
	return
}

func toTime(t float64) time.Time {
	sentSec := int64(t)
	sentNs := int64((t - float64(sentSec)) * 1e9)
	return time.Unix(sentSec, sentNs)
}

func (p *DetectionQualityProcessor) ProcessReferee(_ *persistence.Message, frame *sslproto.Referee) {
	switch *frame.Stage {
	case sslproto.Referee_NORMAL_FIRST_HALF,
		sslproto.Referee_NORMAL_SECOND_HALF,
		sslproto.Referee_EXTRA_FIRST_HALF,
		sslproto.Referee_EXTRA_SECOND_HALF:
	default:
		p.active = false
		return
	}

	switch *frame.Command {
	case sslproto.Referee_HALT,
		sslproto.Referee_TIMEOUT_BLUE,
		sslproto.Referee_TIMEOUT_YELLOW:
		p.active = false
	default:
		p.active = true
	}
}

func (p *DetectionQualityProcessor) String() (res string) {
	res += p.robotSkippedFrames()
	res += p.ballSkippedFrames()
	if p.PrintDataLosses {
		res += p.robotDataLossOverThreshold()
		res += p.ballDataLossOverThreshold()
	} else {
		res += p.robotDataLossOverThresholdSum()
		res += p.ballDataLossOverThresholdSum()
	}
	return
}

func (p *DetectionQualityProcessor) robotDataLossOverThreshold() (res string) {
	res += fmt.Sprintf("Data loss > %v\n", maxRobotLoss)

	sort.Slice(p.robotDataLosses, func(i, j int) bool {
		return p.robotDataLosses[i].Duration < p.robotDataLosses[j].Duration
	})
	for _, dataLoss := range p.robotDataLosses {
		if dataLoss.Duration > maxRobotLoss {
			res += fmt.Sprintf("%42v | %2d %v: %4v frames, %13v (%14v old)\n", dataLoss.Time, dataLoss.RobotId, teamColorStr(dataLoss.TeamColor), dataLoss.NumFrames, dataLoss.Duration, dataLoss.ObjectAge)
		}
	}
	return
}

func (p *DetectionQualityProcessor) ballDataLossOverThreshold() (res string) {
	res += fmt.Sprintf("Data loss > %v\n", maxBallLoss)

	sort.Slice(p.ballDataLosses, func(i, j int) bool {
		return p.ballDataLosses[i].Duration < p.ballDataLosses[j].Duration
	})
	for _, dataLoss := range p.ballDataLosses {
		if dataLoss.Duration > maxBallLoss {
			res += fmt.Sprintf("%42v | ball: %4v frames, %13v (%14v old)\n", dataLoss.Time, dataLoss.NumFrames, dataLoss.Duration, dataLoss.ObjectAge)
		}
	}
	return
}

func (p *DetectionQualityProcessor) robotDataLossOverThresholdSum() (res string) {
	numOverMax := 0
	numOverMaxAndAged := 0
	for _, dataLoss := range p.robotDataLosses {
		if dataLoss.Duration > maxRobotLoss {
			numOverMax++
			if dataLoss.ObjectAge > minRobotAge {
				numOverMaxAndAged++
			}
		}
	}
	res += fmt.Sprintf("Number of robot data losses over %v: %v, %v older than %v\n", maxRobotLoss, numOverMax, numOverMaxAndAged, minRobotAge)
	return
}

func (p *DetectionQualityProcessor) ballDataLossOverThresholdSum() (res string) {
	numOverMax := 0
	numOverMaxAndAged := 0
	for _, dataLoss := range p.ballDataLosses {
		if dataLoss.Duration > maxBallLoss {
			numOverMax++
			if dataLoss.ObjectAge > minBallAge {
				numOverMaxAndAged++
			}
		}
	}
	res += fmt.Sprintf("Number of ball data losses over %v: %v, %v older than %v\n", maxBallLoss, numOverMax, numOverMaxAndAged, minBallAge)
	return
}

func teamColorStr(teamColor TeamColor) string {
	if teamColor == TeamYellow {
		return "Y"
	} else if teamColor == TeamBlue {
		return "B"
	}
	return ""
}

func (p *DetectionQualityProcessor) robotSkippedFrames() (res string) {
	frameMisses := map[int]uint32{}
	for _, dataLoss := range p.robotDataLosses {
		frameMisses[int(dataLoss.NumFrames)]++
	}
	maxNumFrames := 6
	maxNumFramesCount := uint32(0)
	var numFramesList []int
	for numFrames, numFramesCount := range frameMisses {
		if numFrames >= maxNumFrames {
			maxNumFramesCount += numFramesCount
		} else {
			numFramesList = append(numFramesList, numFrames)
		}
	}
	sort.Ints(numFramesList)
	res += fmt.Sprintf("Robots: skipped frames:\n")
	for _, numFrames := range numFramesList {
		res += fmt.Sprintf("%v frames: %vx\n", numFrames, frameMisses[numFrames])
	}
	res += fmt.Sprintf(">=%v frames: %vx\n", maxNumFrames, maxNumFramesCount)
	return res
}

func (p *DetectionQualityProcessor) ballSkippedFrames() (res string) {
	frameMisses := map[int]uint32{}
	for _, dataLoss := range p.ballDataLosses {
		frameMisses[int(dataLoss.NumFrames)]++
	}
	maxNumFrames := 6
	maxNumFramesCount := uint32(0)
	var numFramesList []int
	for numFrames, numFramesCount := range frameMisses {
		if numFrames >= maxNumFrames {
			maxNumFramesCount += numFramesCount
		} else {
			numFramesList = append(numFramesList, numFrames)
		}
	}
	sort.Ints(numFramesList)
	res += fmt.Sprintf("Balls: skipped frames:\n")
	for _, numFrames := range numFramesList {
		res += fmt.Sprintf("%v frames: %vx\n", numFrames, frameMisses[numFrames])
	}
	res += fmt.Sprintf(">=%v frames: %vx\n", maxNumFrames, maxNumFramesCount)
	return res
}
