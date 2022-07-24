package auto

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"strings"
	"time"
)

type Recorder struct {
	server   *sslnet.MulticastServer
	Recorder *persistence.Recorder
}

func NewRecorder(server *sslnet.MulticastServer) (r *Recorder) {
	r = new(Recorder)
	r.server = server
	r.Recorder = new(persistence.Recorder)
	*r.Recorder = persistence.NewRecorder()
	return
}

func (r *Recorder) Start() {
	r.server.Consumer = r.receiveRefereeMessage
	r.server.Start()
}

func (r *Recorder) Stop() {
	r.server.Stop()
	if err := r.Recorder.Stop(); err != nil {
		log.Println("Failed to stop recorder: ", err)
	}
}

func (r *Recorder) receiveRefereeMessage(data []byte, _ *net.UDPAddr) {
	var message referee.Referee
	if err := proto.Unmarshal(data, &message); err != nil {
		log.Println("Could not unmarshal referee message: ", err)
		return
	}

	if !r.Recorder.IsRunning() && isTeamSet(&message) && (isGameStage(&message) || isPreGameStage(&message)) {
		name := logFileName(&message)
		log.Println("Start recording ", name)
		if err := r.Recorder.StartWithName(name); err != nil {
			log.Println("Failed to start recorder: ", err)
		}
	} else if r.Recorder.IsRunning() {
		if isPostGame(&message) || !isTeamSet(&message) {
			log.Println("Stop recording")
			if err := r.Recorder.Stop(); err != nil {
				log.Println("Failed to stop recorder: ", err)
			}
		} else if !r.Recorder.Paused && isBreakStage(&message) {
			log.Println("Pause recording")
			r.Recorder.Paused = true
		} else if r.Recorder.Paused && !isBreakStage(&message) {
			log.Println("Resume recording")
			r.Recorder.Paused = false
		}
	}
}

func logFileName(refMsg *referee.Referee) string {
	teamNameYellow := strings.Replace(*refMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*refMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	return fmt.Sprintf("%s_%s-vs-%s", date, teamNameYellow, teamNameBlue)
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
