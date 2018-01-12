package main

import (
	"time"
	"log"
	"io/ioutil"
	"fmt"
	"os"
)

type GameState int

const (
	GameStateHalt    GameState = iota
	GameStateStop
	GameStateRunning
)

type MatchStats struct {
	Filename       string
	MatchTotalTime time.Duration
	StateTotalTime map[GameState]time.Duration
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Please pass a directory to analyse")
	}
	dir :=  os.Args[1]

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		m, err := matchStats(dir + "/" + f.Name())
		if err != nil {
			fmt.Printf("%v: %v\n", m.Filename, err)
		} else {
			checkSum := m.MatchTotalTime - (m.StateTotalTime[0] + m.StateTotalTime[1] + m.StateTotalTime[2])
			fmt.Printf("%v,%f,%f,%f,%f,%f\n", m.Filename, m.MatchTotalTime.Seconds(), m.StateTotalTime[0].Seconds(), m.StateTotalTime[1].Seconds(), m.StateTotalTime[2].Seconds(), checkSum.Seconds())
		}
	}
}

func matchStats(filename string) (m *MatchStats, err error) {
	m = new(MatchStats)
	m.Filename = filename

	logReader, err := NewLogReader(filename)
	if err != nil {
		return
	}
	defer logReader.Close()

	channel := make(chan *SSL_Referee, 100)
	go logReader.CreateRefereeChannel(channel)

	var lastCmdId *uint32
	var halfStartTimestamp *uint64
	gameState := GameStateHalt
	var lastGameStateTimestamp *uint64
	m.StateTotalTime = make(map[GameState]time.Duration)
	m.StateTotalTime[GameStateHalt] = 0
	m.StateTotalTime[GameStateStop] = 0
	m.StateTotalTime[GameStateRunning] = 0

	var lastRefereeMsg *SSL_Referee
	for r := range channel {
		if lastCmdId != nil && *r.CommandCounter == *lastCmdId {
			continue
		}
		lastCmdId = r.CommandCounter
		switch *r.Stage {
		case SSL_Referee_NORMAL_FIRST_HALF, SSL_Referee_NORMAL_SECOND_HALF, SSL_Referee_EXTRA_FIRST_HALF, SSL_Referee_EXTRA_SECOND_HALF:
			if halfStartTimestamp == nil {
				halfStartTimestamp = r.CommandTimestamp
			}
			if lastGameStateTimestamp == nil {
				lastGameStateTimestamp = r.CommandTimestamp
			}
			var nextGameState GameState
			switch *r.Command {
			case SSL_Referee_HALT, SSL_Referee_TIMEOUT_BLUE, SSL_Referee_TIMEOUT_YELLOW:
				nextGameState = GameStateHalt
			case SSL_Referee_STOP, SSL_Referee_BALL_PLACEMENT_BLUE, SSL_Referee_BALL_PLACEMENT_YELLOW:
				nextGameState = GameStateStop
			case SSL_Referee_FORCE_START, SSL_Referee_NORMAL_START, SSL_Referee_DIRECT_FREE_YELLOW, SSL_Referee_DIRECT_FREE_BLUE, SSL_Referee_INDIRECT_FREE_YELLOW, SSL_Referee_INDIRECT_FREE_BLUE, SSL_Referee_PREPARE_KICKOFF_BLUE, SSL_Referee_PREPARE_KICKOFF_YELLOW, SSL_Referee_PREPARE_PENALTY_BLUE, SSL_Referee_PREPARE_PENALTY_YELLOW:
				nextGameState = GameStateRunning
			case SSL_Referee_GOAL_BLUE, SSL_Referee_GOAL_YELLOW:
				nextGameState = gameState
			default:
				log.Printf("Unknown command: %v", *r.Command)
			}
			if nextGameState != gameState {
				m.StateTotalTime[gameState] += time.Duration((*r.CommandTimestamp - *lastGameStateTimestamp) * 1000)
				gameState = nextGameState
				lastGameStateTimestamp = r.CommandTimestamp
			}
		case SSL_Referee_NORMAL_HALF_TIME, SSL_Referee_EXTRA_TIME_BREAK, SSL_Referee_EXTRA_HALF_TIME, SSL_Referee_POST_GAME:
			if halfStartTimestamp != nil {
				m.MatchTotalTime += time.Duration((*r.CommandTimestamp - *halfStartTimestamp) * 1000)
				halfStartTimestamp = nil
			}
			if lastGameStateTimestamp != nil {
				m.StateTotalTime[gameState] += time.Duration((*r.CommandTimestamp - *lastGameStateTimestamp) * 1000)
			}
			lastGameStateTimestamp = nil
		}
		lastRefereeMsg = r
	}
	if halfStartTimestamp != nil && lastRefereeMsg != nil {
		m.MatchTotalTime += time.Duration((*lastRefereeMsg.CommandTimestamp - *halfStartTimestamp) * 1000)
	}
	return
}
