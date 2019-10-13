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
	matchStats.TeamStatsBlue.Name = *referee.Blue.Name
	matchStats.TeamStatsYellow.Name = *referee.Yellow.Name
	endTime := packetTimeStampToTime(*referee.PacketTimestamp)
	matchStats.MatchDuration = float32(endTime.Sub(m.startTime).Seconds())
}

func packetTimeStampToTime(packetTimestamp uint64) time.Time {
	seconds := int64(packetTimestamp / 1_000_000)
	nanoSeconds := int64(packetTimestamp-uint64(seconds*1_000_000)) * 1000
	return time.Unix(seconds, nanoSeconds)
}
