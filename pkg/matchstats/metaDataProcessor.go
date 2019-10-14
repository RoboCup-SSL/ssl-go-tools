package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"time"
)

type MetaDataProcessor struct {
	startTime         time.Time
	timeoutTimeNormal uint32
	timeoutsNormal    uint32
	timeoutTimeExtra  uint32
	timeoutsExtra     uint32
}

func (m *MetaDataProcessor) OnNewStage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	if *referee.Stage == sslproto.Referee_EXTRA_TIME_BREAK {
		matchStats.TeamStatsYellow.TimeoutTime = m.timeoutTimeNormal - *referee.Yellow.TimeoutTime
		matchStats.TeamStatsBlue.TimeoutTime = m.timeoutTimeNormal - *referee.Blue.TimeoutTime
		matchStats.TeamStatsYellow.Timeouts = m.timeoutsNormal - *referee.Yellow.Timeouts
		matchStats.TeamStatsBlue.Timeouts = m.timeoutsNormal - *referee.Blue.Timeouts
	}
	if *referee.Stage == sslproto.Referee_PENALTY_SHOOTOUT {
		matchStats.Shootout = true
	}
}

func (m *MetaDataProcessor) OnNewCommand(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	switch *referee.Command {
	case sslproto.Referee_PREPARE_PENALTY_BLUE:
		matchStats.TeamStatsBlue.PenaltyShotsTotal += 1
	case sslproto.Referee_PREPARE_PENALTY_YELLOW:
		matchStats.TeamStatsYellow.PenaltyShotsTotal += 1
	}
}

func (m *MetaDataProcessor) OnFirstRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	if matchStats.TeamStatsBlue == nil {
		matchStats.TeamStatsBlue = new(sslproto.TeamStats)
	}
	if matchStats.TeamStatsYellow == nil {
		matchStats.TeamStatsYellow = new(sslproto.TeamStats)
	}

	matchStats.Shootout = false
	m.startTime = packetTimeStampToTime(*referee.PacketTimestamp)
}

func (m *MetaDataProcessor) OnLastRefereeMessage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	processTeam(matchStats.TeamStatsBlue, referee.Blue)
	processTeam(matchStats.TeamStatsYellow, referee.Yellow)
	endTime := packetTimeStampToTime(*referee.PacketTimestamp)
	matchStats.MatchDuration = uint32(endTime.Sub(m.startTime).Microseconds())

	if uint32(*referee.Stage) <= uint32(sslproto.Referee_NORMAL_SECOND_HALF) {
		matchStats.TeamStatsYellow.TimeoutTime = m.timeoutTimeNormal - *referee.Yellow.TimeoutTime
		matchStats.TeamStatsBlue.TimeoutTime = m.timeoutTimeNormal - *referee.Blue.TimeoutTime
		matchStats.TeamStatsYellow.Timeouts = m.timeoutsNormal - *referee.Yellow.Timeouts
		matchStats.TeamStatsBlue.Timeouts = m.timeoutsNormal - *referee.Blue.Timeouts
		matchStats.ExtraTime = false
	} else {
		matchStats.TeamStatsYellow.TimeoutTime += m.timeoutTimeExtra - *referee.Yellow.TimeoutTime
		matchStats.TeamStatsBlue.TimeoutTime += m.timeoutTimeExtra - *referee.Blue.TimeoutTime
		matchStats.TeamStatsYellow.Timeouts += m.timeoutsExtra - *referee.Yellow.Timeouts
		matchStats.TeamStatsBlue.Timeouts += m.timeoutsExtra - *referee.Blue.Timeouts
		matchStats.ExtraTime = true
	}
}

func processTeam(stats *sslproto.TeamStats, team *sslproto.Referee_TeamInfo) {
	stats.Name = *team.Name
	stats.Goals = *team.Score
	stats.Fouls = *team.FoulCounter
	stats.YellowCards = *team.YellowCards
	stats.RedCards = *team.RedCards
}
