package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/pkg/errors"
	"log"
)

type MatchStatProcessor interface {
	OnNewStage(referee *sslproto.Referee)
	OnNewCommand(referee *sslproto.Referee)
	OnNewGameEvent(referee *sslproto.Referee)
}

type MatchStatGenerator struct {
	processors []*MatchStatProcessor
}

func NewMatchStatsGenerator() *MatchStatGenerator {
	return new(MatchStatGenerator)
}

func (m *MatchStatGenerator) Process(filename string) (*sslproto.MatchStats, error) {

	logReader, err := persistence.NewReader(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read file")
	}

	matchStats := new(sslproto.MatchStats)
	var lastRefereeMsg *sslproto.Referee

	channel := logReader.CreateChannel()
	for c := range channel {
		if c.MessageType.Id != persistence.MessageSslRefbox2013 {
			continue
		}
		r, err := c.ParseReferee()
		if err != nil {
			log.Println("Could not parse referee message: ", err)
			continue
		}
		if lastRefereeMsg == nil || *r.Command != *lastRefereeMsg.Command {
			for _, p := range m.processors {
				(*p).OnNewCommand(r)
			}
		}

		if lastRefereeMsg == nil || *r.Stage != *lastRefereeMsg.Stage {
			for _, p := range m.processors {
				(*p).OnNewStage(r)
			}
		}

		//if lastRefereeMsg == nil || *r.GameEvents != *lastRefereeMsg.GameEvents {
		//	for _, p := range m.processors {
		//		(*p).OnNewStage(r)
		//	}
		//}

		lastRefereeMsg = r
	}

	return matchStats, logReader.Close()
}
