package auto

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
)

type Recorder struct {
	server   *MulticastServer
	Recorder *persistence.Recorder
}

func NewRecorder(server *MulticastServer) (r *Recorder) {
	r = new(Recorder)
	r.server = server
	r.Recorder = new(persistence.Recorder)
	*r.Recorder = persistence.NewRecorder()
	return
}

func (r *Recorder) Start() {
	r.server.consumer = r.receiveRefereeMessage
	r.server.Start()
}

func (r *Recorder) Stop() {
	r.server.Stop()
	if err := r.Recorder.Stop(); err != nil {
		log.Println("Failed to stop recorder: ", err)
	}
}

func (r *Recorder) receiveRefereeMessage(data []byte, _ *net.UDPAddr) {
	var message sslproto.Referee
	if err := proto.Unmarshal(data, &message); err != nil {
		log.Println("Could not unmarshal referee message: ", err)
		return
	}

	if !r.Recorder.IsRunning() && (isGameStage(&message) || isPreGameStage(&message)) {
		log.Println("Start recording")
		if err := r.Recorder.Start(); err != nil {
			log.Println("Failed to start recorder: ", err)
		}
	} else if r.Recorder.IsRunning() && isNoGameStage(&message) {
		log.Println("Stop recording")
		if err := r.Recorder.Stop(); err != nil {
			log.Println("Failed to stop recorder: ", err)
		}
	}
}

func isGameStage(message *sslproto.Referee) bool {
	switch *message.Stage {
	case sslproto.Referee_NORMAL_FIRST_HALF,
		sslproto.Referee_NORMAL_SECOND_HALF,
		sslproto.Referee_EXTRA_FIRST_HALF,
		sslproto.Referee_EXTRA_SECOND_HALF,
		sslproto.Referee_PENALTY_SHOOTOUT:
		return true
	default:
		return false
	}
}

func isNoGameStage(message *sslproto.Referee) bool {
	switch *message.Stage {
	case sslproto.Referee_EXTRA_HALF_TIME,
		sslproto.Referee_NORMAL_HALF_TIME,
		sslproto.Referee_PENALTY_SHOOTOUT_BREAK,
		sslproto.Referee_POST_GAME,
		sslproto.Referee_EXTRA_TIME_BREAK:
		return true
	default:
		return false
	}
}

func isPreStage(message *sslproto.Referee) bool {
	switch *message.Stage {
	case sslproto.Referee_NORMAL_FIRST_HALF_PRE,
		sslproto.Referee_NORMAL_SECOND_HALF_PRE,
		sslproto.Referee_EXTRA_FIRST_HALF_PRE,
		sslproto.Referee_EXTRA_SECOND_HALF_PRE:
		return true
	default:
		return false
	}
}

func isPreGameStage(message *sslproto.Referee) bool {
	return isPreStage(message) && *message.Command != sslproto.Referee_HALT
}
