package auto

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sourcefilter"
	"google.golang.org/protobuf/proto"
)

type Recorder struct {
	Recorder     *persistence.Recorder
	logFilePath  string
	logFileDir   string
	sourceFilter *sourcefilter.SourceFilter
}

func NewRecorder(logFileDir string) (r *Recorder) {
	r = new(Recorder)
	r.Recorder = new(persistence.Recorder)
	*r.Recorder = persistence.NewRecorder()
	r.Recorder.AddMessageConsumer(r.consumeMessage)
	r.logFileDir = logFileDir
	if err := os.MkdirAll(r.logFileDir, os.ModePerm); err != nil {
		log.Println("Could not create log file dir", err)
	}
	return
}

func (r *Recorder) SetSourceFilter(filter *sourcefilter.SourceFilter) {
	r.sourceFilter = filter
}

func (r *Recorder) Start() {
	r.Recorder.StartReceiving()
}

func (r *Recorder) Stop() {
	r.Recorder.StopReceiving()
	r.StopRecording()
}

func (r *Recorder) DiscardRecording() {
	log.Println("Discard recording")
	if err := r.Recorder.StopRecording(); err != nil {
		log.Println("Failed to stop recorder: ", err)
	}
	if err := os.Remove(r.logFilePath); err != nil {
		log.Println("Could not remove log file:", r.logFilePath, err)
	}
}

func (r *Recorder) StopRecording() {
	log.Println("Stop recording")
	if err := r.Recorder.StopRecording(); err != nil {
		log.Println("Failed to stop recorder: ", err)
	}
	if err := index.WriteIndex(r.logFilePath); err != nil {
		log.Println("Could not index log file:", r.logFilePath, err)
	}
	if err := persistence.Compress(r.logFilePath, r.logFilePath+".gz"); err != nil {
		log.Println("Could not compress log file:", r.logFilePath, err)
	}
}

func (r *Recorder) consumeMessage(message *persistence.Message, addr *net.UDPAddr) {
	if message.MessageType.Id != persistence.MessageSslRefbox2013 {
		return
	}

	// Apply source filter for referee messages
	if r.sourceFilter != nil && addr != nil {
		if !r.sourceFilter.Accept(addr.IP) {
			return // Reject message from non-active source
		}
	}

	var refMsg gc.Referee

	if err := proto.Unmarshal(message.Message, &refMsg); err != nil {
		log.Println("Could not unmarshal referee message: ", err)
		return
	}

	if !r.Recorder.IsRecording() && isTeamSet(&refMsg) &&
		*refMsg.Command != gc.Referee_HALT &&
		!isPostGame(&refMsg) {
		logFileName := LogFileName(&refMsg, time.UTC)
		r.logFilePath = filepath.Join(r.logFileDir, logFileName)
		log.Println("Start recording ", r.logFilePath)
		if err := r.Recorder.StartRecording(r.logFilePath); err != nil {
			log.Println("Failed to start recorder: ", err)
		}
	} else if r.Recorder.IsRecording() {
		if isPostGame(&refMsg) || !isTeamSet(&refMsg) {
			r.StopRecording()
		} else if *refMsg.Command == gc.Referee_HALT && *refMsg.Stage == gc.Referee_NORMAL_FIRST_HALF_PRE {
			r.DiscardRecording()
		} else if !r.Recorder.IsPaused() && isBreakStage(&refMsg) {
			log.Println("Pause recording")
			r.Recorder.SetPaused(true)
		} else if r.Recorder.IsPaused() && !isBreakStage(&refMsg) {
			log.Println("Resume recording")
			r.Recorder.SetPaused(false)
		}
	}
}

func isTeamSet(message *gc.Referee) bool {
	return *message.Blue.Name != "Unknown" && *message.Yellow.Name != "Unknown" &&
		*message.Blue.Name != "" && *message.Yellow.Name != ""
}

func isBreakStage(message *gc.Referee) bool {
	switch *message.Stage {
	case gc.Referee_EXTRA_HALF_TIME,
		gc.Referee_NORMAL_HALF_TIME,
		gc.Referee_PENALTY_SHOOTOUT_BREAK,
		gc.Referee_EXTRA_TIME_BREAK:
		return true
	default:
		return false
	}
}

func isPostGame(message *gc.Referee) bool {
	return *message.Stage == gc.Referee_POST_GAME
}
