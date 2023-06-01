package auto

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/index"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Recorder struct {
	Recorder    *persistence.Recorder
	logFileName string
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
	if err := index.WriteIndex(r.logFileName); err != nil {
		log.Println("Could not index log file:", r.logFileName, err)
	}
	if err := os.Rename(r.logFileName, filepath.Join(r.logFileDir, r.logFileName)); err != nil {
		log.Println("Could not move log file", err)
	}
}

func (r *Recorder) consumeMessage(message *persistence.Message) {
	if message.MessageType.Id != persistence.MessageSslRefbox2013 {
		return
	}
	var refMsg referee.Referee

	if err := proto.Unmarshal(message.Message, &refMsg); err != nil {
		log.Println("Could not unmarshal referee message: ", err)
		return
	}

	if !r.Recorder.IsRecording() && isTeamSet(&refMsg) && (isGameStage(&refMsg) || isPreGameStage(&refMsg)) {
		r.logFileName = logFileName(&refMsg)
		log.Println("Start recording ", r.logFileName)
		if err := r.Recorder.StartRecording(r.logFileName); err != nil {
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

func logFileName(refMsg *referee.Referee) string {
	teamNameYellow := strings.Replace(*refMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*refMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	return fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
}

func isGameStage(message *referee.Referee) bool {
	switch *message.Stage {
	case referee.Referee_NORMAL_FIRST_HALF,
		referee.Referee_NORMAL_SECOND_HALF,
		referee.Referee_EXTRA_FIRST_HALF,
		referee.Referee_EXTRA_SECOND_HALF,
		referee.Referee_PENALTY_SHOOTOUT:
		return true
	default:
		return false
	}
}

func isTeamSet(message *referee.Referee) bool {
	return *message.Blue.Name != "Unknown" && *message.Yellow.Name != "Unknown" &&
		*message.Blue.Name != "" && *message.Yellow.Name != ""
}

func isBreakStage(message *referee.Referee) bool {
	switch *message.Stage {
	case referee.Referee_EXTRA_HALF_TIME,
		referee.Referee_NORMAL_HALF_TIME,
		referee.Referee_PENALTY_SHOOTOUT_BREAK,
		referee.Referee_EXTRA_TIME_BREAK:
		return true
	default:
		return false
	}
}

func isPostGame(message *referee.Referee) bool {
	return *message.Stage == referee.Referee_POST_GAME
}

func isPreStage(message *referee.Referee) bool {
	switch *message.Stage {
	case referee.Referee_NORMAL_FIRST_HALF_PRE,
		referee.Referee_NORMAL_SECOND_HALF_PRE,
		referee.Referee_EXTRA_FIRST_HALF_PRE,
		referee.Referee_EXTRA_SECOND_HALF_PRE:
		return true
	default:
		return false
	}
}

func isPreGameStage(message *referee.Referee) bool {
	return isPreStage(message) && *message.Command != referee.Referee_HALT
}
