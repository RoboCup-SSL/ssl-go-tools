package auto

import (
	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Recorder struct {
	Recorder    *persistence.Recorder
	logFilePath string
	logFileDir  string
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

func (r *Recorder) Start() {
	r.Recorder.StartReceiving()
}

func (r *Recorder) Stop() {
	r.Recorder.StopReceiving()
	r.StopRecording()
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

func (r *Recorder) consumeMessage(message *persistence.Message) {
	if message.MessageType.Id != persistence.MessageSslRefbox2013 {
		return
	}
	var refMsg gc.Referee

	if err := proto.Unmarshal(message.Message, &refMsg); err != nil {
		log.Println("Could not unmarshal referee message: ", err)
		return
	}

	if !r.Recorder.IsRecording() && isTeamSet(&refMsg) && (isGameStage(&refMsg) || isPreGameStage(&refMsg)) {
		logFileName := LogFileName(&refMsg, time.UTC)
		r.logFilePath = filepath.Join(r.logFileDir, logFileName)
		log.Println("Start recording ", r.logFilePath)
		if err := r.Recorder.StartRecording(r.logFilePath); err != nil {
			log.Println("Failed to start recorder: ", err)
		}
	} else if r.Recorder.IsRecording() {
		if isPostGame(&refMsg) || !isTeamSet(&refMsg) {
			r.StopRecording()
		} else if !r.Recorder.IsPaused() && isBreakStage(&refMsg) {
			log.Println("Pause recording")
			r.Recorder.SetPaused(true)
		} else if r.Recorder.IsPaused() && !isBreakStage(&refMsg) {
			log.Println("Resume recording")
			r.Recorder.SetPaused(false)
		}
	}
}

func isGameStage(message *gc.Referee) bool {
	switch *message.Stage {
	case gc.Referee_NORMAL_FIRST_HALF,
		gc.Referee_NORMAL_SECOND_HALF,
		gc.Referee_EXTRA_FIRST_HALF,
		gc.Referee_EXTRA_SECOND_HALF,
		gc.Referee_PENALTY_SHOOTOUT:
		return true
	default:
		return false
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

func isPreStage(message *gc.Referee) bool {
	switch *message.Stage {
	case gc.Referee_NORMAL_FIRST_HALF_PRE,
		gc.Referee_NORMAL_SECOND_HALF_PRE,
		gc.Referee_EXTRA_FIRST_HALF_PRE,
		gc.Referee_EXTRA_SECOND_HALF_PRE:
		return true
	default:
		return false
	}
}

func isPreGameStage(message *gc.Referee) bool {
	return isPreStage(message) && *message.Command != gc.Referee_HALT
}
