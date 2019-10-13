package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"log"
)

type GamePhaseAggregator struct {
}

func (d *GamePhaseAggregator) Aggregate(matchStats *sslproto.MatchStats) {

	matchStats.TimePerGamePhase = map[string]uint32{}
	matchStats.RelTimePerGamePhase = map[string]float32{}

	for _, p := range sslproto.GamePhaseType_name {
		matchStats.TimePerGamePhase[p] = 0
	}

	for _, p := range matchStats.GamePhases {
		matchStats.TimePerGamePhase[(*p).Type.String()] += p.Duration
	}

	checkSum := uint32(0)
	for _, p := range sslproto.GamePhaseType_name {
		checkSum += matchStats.TimePerGamePhase[p]
		matchStats.RelTimePerGamePhase[p] = float32(matchStats.TimePerGamePhase[p]) / float32(matchStats.MatchDuration)
	}

	if matchStats.MatchDuration != checkSum {
		log.Printf("Match duration mismatch. Total: %v, Sum of phases: %v, Diff: %v", matchStats.MatchDuration, checkSum, matchStats.MatchDuration-checkSum)
	}
}
