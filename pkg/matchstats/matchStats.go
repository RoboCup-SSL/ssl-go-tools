package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/pkg/errors"
	"log"
	"path/filepath"
)

type MatchStatGenerator struct {
	metaDataProcessor MetaDataProcessor
}

func NewMatchStatsGenerator() *MatchStatGenerator {
	generator := new(MatchStatGenerator)
	generator.metaDataProcessor = MetaDataProcessor{}
	return generator
}

func (m *MatchStatGenerator) Process(filename string) (*sslproto.MatchStats, error) {

	logReader, err := persistence.NewReader(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read file")
	}

	matchStats := new(sslproto.MatchStats)
	matchStats.Name = filepath.Base(filename)
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

		if lastRefereeMsg == nil {
			m.OnFirstRefereeMessage(matchStats, r)
		}

		if lastRefereeMsg == nil || *r.Command != *lastRefereeMsg.Command {
			m.OnNewCommand(matchStats, r)
		}

		if lastRefereeMsg == nil || *r.Stage != *lastRefereeMsg.Stage {
			m.OnNewStage(matchStats, r)
		}

		lastRefereeMsg = r
	}

	m.OnLastRefereeMessage(matchStats, lastRefereeMsg)

	return matchStats, logReader.Close()
}

func (m *MatchStatGenerator) OnNewStage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {

}

func (m *MatchStatGenerator) OnNewCommand(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {

}

func (m *MatchStatGenerator) OnNewGameEvent(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {

}

func (m *MatchStatGenerator) OnFirstRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	m.metaDataProcessor.OnFirstRefereeMessage(matchStats, referee)
}

func (m *MatchStatGenerator) OnLastRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	m.metaDataProcessor.OnLastRefereeMessage(matchStats, referee)
}
