package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"time"
)

type MetaDataProcessor struct {
	startTime time.Time
}

func (m *MetaDataProcessor) OnFirstRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	m.startTime = packetTimeStampToTime(*referee.PacketTimestamp)
}

func (m *MetaDataProcessor) OnLastRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	if matchStats.TeamStatsBlue == nil {
		matchStats.TeamStatsBlue = new(sslproto.TeamStats)
	}
	if matchStats.TeamStatsYellow == nil {
		matchStats.TeamStatsYellow = new(sslproto.TeamStats)
	}
	processTeam(matchStats.TeamStatsBlue, referee.Blue)
	processTeam(matchStats.TeamStatsYellow, referee.Yellow)
	endTime := packetTimeStampToTime(*referee.PacketTimestamp)
	matchStats.MatchDuration = float32(endTime.Sub(m.startTime).Seconds())
}

func processTeam(stats *sslproto.TeamStats, team *sslproto.Referee_TeamInfo) {
	stats.Name = *team.Name
	stats.Goals = *team.Score
	stats.Fouls = *team.FoulCounter
	stats.YellowCards = *team.YellowCards
	stats.RedCards = *team.RedCards
}
