package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"log"
)

type GamePhaseDetector struct {
	currentPhase *sslproto.GamePhase
	gamePaused   bool
}

func (d *GamePhaseDetector) startNewPhase(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	if d.currentPhase != nil {
		d.currentPhase.EndTime = *referee.PacketTimestamp
		start := packetTimeStampToTime(d.currentPhase.StartTime)
		end := packetTimeStampToTime(d.currentPhase.EndTime)
		d.currentPhase.Duration = float32(end.Sub(start).Seconds())
		matchStats.GamePhases = append(matchStats.GamePhases, d.currentPhase)
		d.currentPhase.CommandExit = mapProtoCommandToCommand(*referee.Command)
		if referee.NextCommand != nil && int32(*referee.NextCommand) >= 0 {
			d.currentPhase.NextCommandProposed = mapProtoCommandToCommand(*referee.NextCommand)
		}
		d.currentPhase.GameEventsExit = referee.GameEvents
	}
	d.currentPhase = new(sslproto.GamePhase)
	d.currentPhase.StartTime = *referee.PacketTimestamp
}

func (d *GamePhaseDetector) OnNewStage(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	switch *referee.Stage {
	case sslproto.Referee_NORMAL_FIRST_HALF_PRE,
		sslproto.Referee_NORMAL_FIRST_HALF,
		sslproto.Referee_NORMAL_SECOND_HALF_PRE,
		sslproto.Referee_NORMAL_SECOND_HALF,
		sslproto.Referee_EXTRA_FIRST_HALF_PRE,
		sslproto.Referee_EXTRA_SECOND_HALF_PRE,
		sslproto.Referee_EXTRA_FIRST_HALF,
		sslproto.Referee_EXTRA_SECOND_HALF,
		sslproto.Referee_PENALTY_SHOOTOUT:
		d.gamePaused = false
		break
	case sslproto.Referee_NORMAL_HALF_TIME,
		sslproto.Referee_EXTRA_TIME_BREAK,
		sslproto.Referee_EXTRA_HALF_TIME,
		sslproto.Referee_PENALTY_SHOOTOUT_BREAK:
		d.gamePaused = true
		d.startNewPhase(matchStats, referee)
		d.currentPhase.Type = sslproto.GamePhaseType_PHASE_BREAK
		break
	case sslproto.Referee_POST_GAME:
		d.startNewPhase(matchStats, referee)
		d.currentPhase.Type = sslproto.GamePhaseType_PHASE_POST_GAME
	}
}

func (d *GamePhaseDetector) OnNewCommand(matchStats *sslproto.MatchStats, referee *sslproto.Referee) {
	if d.gamePaused {
		return
	}

	d.startNewPhase(matchStats, referee)
	d.currentPhase.Type = mapProtoCommandToGamePhaseType(*referee.Command)
	d.currentPhase.CommandEntry = mapProtoCommandToCommand(*referee.Command)
	d.currentPhase.ForTeam = mapProtoCommandToTeam(*referee.Command)
	d.currentPhase.GameEventsEntry = referee.GameEvents
}

func mapProtoCommandToCommand(command sslproto.Referee_Command) *sslproto.Command {
	return &sslproto.Command{
		Type:    mapProtoCommandToCommandType(command),
		ForTeam: mapProtoCommandToTeam(command),
	}
}

func mapProtoCommandToCommandType(command sslproto.Referee_Command) sslproto.CommandType {
	switch command {
	case sslproto.Referee_HALT:
		return sslproto.CommandType_COMMAND_HALT
	case sslproto.Referee_STOP:
		return sslproto.CommandType_COMMAND_STOP
	case sslproto.Referee_NORMAL_START:
		return sslproto.CommandType_COMMAND_NORMAL_START
	case sslproto.Referee_FORCE_START:
		return sslproto.CommandType_COMMAND_FORCE_START
	case sslproto.Referee_PREPARE_KICKOFF_YELLOW,
		sslproto.Referee_PREPARE_KICKOFF_BLUE:
		return sslproto.CommandType_COMMAND_PREPARE_KICKOFF
	case sslproto.Referee_PREPARE_PENALTY_YELLOW,
		sslproto.Referee_PREPARE_PENALTY_BLUE:
		return sslproto.CommandType_COMMAND_PREPARE_PENALTY
	case sslproto.Referee_DIRECT_FREE_YELLOW,
		sslproto.Referee_DIRECT_FREE_BLUE:
		return sslproto.CommandType_COMMAND_DIRECT_FREE
	case sslproto.Referee_INDIRECT_FREE_YELLOW,
		sslproto.Referee_INDIRECT_FREE_BLUE:
		return sslproto.CommandType_COMMAND_INDIRECT_FREE
	case sslproto.Referee_TIMEOUT_YELLOW,
		sslproto.Referee_TIMEOUT_BLUE:
		return sslproto.CommandType_COMMAND_TIMEOUT
	case sslproto.Referee_BALL_PLACEMENT_YELLOW,
		sslproto.Referee_BALL_PLACEMENT_BLUE:
		return sslproto.CommandType_COMMAND_BALL_PLACEMENT
	case sslproto.Referee_GOAL_YELLOW, sslproto.Referee_GOAL_BLUE:
		return sslproto.CommandType_COMMAND_UNKNOWN
	}
	log.Printf("Command %v not mapped to any command type", command)
	return sslproto.CommandType_COMMAND_UNKNOWN
}

func mapProtoCommandToGamePhaseType(command sslproto.Referee_Command) sslproto.GamePhaseType {
	switch command {
	case sslproto.Referee_HALT:
		return sslproto.GamePhaseType_PHASE_HALT
	case sslproto.Referee_STOP:
		return sslproto.GamePhaseType_PHASE_STOP
	case sslproto.Referee_NORMAL_START:
		return sslproto.GamePhaseType_PHASE_RUNNING
	case sslproto.Referee_FORCE_START:
		return sslproto.GamePhaseType_PHASE_RUNNING
	case sslproto.Referee_PREPARE_KICKOFF_YELLOW, sslproto.Referee_PREPARE_KICKOFF_BLUE:
		return sslproto.GamePhaseType_PHASE_PREPARE_KICKOFF
	case sslproto.Referee_PREPARE_PENALTY_YELLOW, sslproto.Referee_PREPARE_PENALTY_BLUE:
		return sslproto.GamePhaseType_PHASE_PREPARE_PENALTY
	case sslproto.Referee_DIRECT_FREE_YELLOW, sslproto.Referee_DIRECT_FREE_BLUE:
		return sslproto.GamePhaseType_PHASE_DIRECT_FREE
	case sslproto.Referee_INDIRECT_FREE_YELLOW, sslproto.Referee_INDIRECT_FREE_BLUE:
		return sslproto.GamePhaseType_PHASE_INDIRECT_FREE
	case sslproto.Referee_TIMEOUT_YELLOW, sslproto.Referee_TIMEOUT_BLUE:
		return sslproto.GamePhaseType_PHASE_TIMEOUT
	case sslproto.Referee_BALL_PLACEMENT_YELLOW, sslproto.Referee_BALL_PLACEMENT_BLUE:
		return sslproto.GamePhaseType_PHASE_BALL_PLACEMENT
	case sslproto.Referee_GOAL_YELLOW, sslproto.Referee_GOAL_BLUE:
		return sslproto.GamePhaseType_PHASE_UNKNOWN
	}
	log.Printf("Command %v not mapped to any phase type", command)
	return sslproto.GamePhaseType_PHASE_UNKNOWN
}

func mapProtoCommandToTeam(command sslproto.Referee_Command) sslproto.TeamColor {
	switch command {
	case sslproto.Referee_HALT,
		sslproto.Referee_STOP,
		sslproto.Referee_NORMAL_START,
		sslproto.Referee_FORCE_START,
		sslproto.Referee_GOAL_YELLOW,
		sslproto.Referee_GOAL_BLUE:
		return sslproto.TeamColor_TEAM_UNKNOWN
	case sslproto.Referee_PREPARE_KICKOFF_YELLOW,
		sslproto.Referee_PREPARE_PENALTY_YELLOW,
		sslproto.Referee_DIRECT_FREE_YELLOW,
		sslproto.Referee_INDIRECT_FREE_YELLOW,
		sslproto.Referee_TIMEOUT_YELLOW,
		sslproto.Referee_BALL_PLACEMENT_YELLOW:
		return sslproto.TeamColor_TEAM_YELLOW
	case sslproto.Referee_PREPARE_KICKOFF_BLUE,
		sslproto.Referee_PREPARE_PENALTY_BLUE,
		sslproto.Referee_DIRECT_FREE_BLUE,
		sslproto.Referee_INDIRECT_FREE_BLUE,
		sslproto.Referee_TIMEOUT_BLUE,
		sslproto.Referee_BALL_PLACEMENT_BLUE:
		return sslproto.TeamColor_TEAM_BLUE
	}
	log.Printf("Command %v not mapped to any team", command)
	return sslproto.TeamColor_TEAM_UNKNOWN
}