package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"strings"
	"time"
)

const tmpLogFilename = "tmp.log.gz"

var logCutter LogCutter

type LogCutter struct {
	logWriter       *persistence.Writer
	firstRefereeMsg *referee.Referee
	lastRefereeMsg  *referee.Referee
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	for _, arg := range args {
		log.Println("Processing", arg)
		process(arg)
		log.Println("Processing done")
		log.Println("")
	}

	logCutter.Stop()
}

func (l *LogCutter) Running() bool {
	return l.logWriter != nil
}

func (l *LogCutter) Stopped() bool {
	return l.logWriter == nil
}

func (l *LogCutter) Start() {
	if l.Running() {
		return
	}
	log.Print("Start log writer")
	if logWriter, err := persistence.NewWriter(tmpLogFilename); err != nil {
		log.Fatal("Could not open temporary writer:", err)
	} else {
		l.logWriter = logWriter
	}
}

func (l *LogCutter) Update(refereeMsg *referee.Referee) {
	if l.Stopped() {
		return
	}
	if l.firstRefereeMsg == nil {
		l.firstRefereeMsg = refereeMsg
	}
	l.lastRefereeMsg = refereeMsg
}

func (l *LogCutter) Write(logMessage *persistence.Message) {
	if l.Stopped() {
		return
	}
	if err := l.logWriter.Write(logMessage); err != nil {
		log.Println("Could not write log message:", err)
	}
}

func usage() {
	_, err := fmt.Fprintln(os.Stderr, "Pass one or more log files that should be cut.")
	if err != nil {
		fmt.Println("Pass one or more log files that should be cut.")
	}
	flag.PrintDefaults()
}

func process(filename string) {
	logReader, err := persistence.NewReader(filename)
	if err != nil {
		log.Println("Could not process log file:", err)
		return
	}

	channel := logReader.CreateChannel()

	var lastStage *referee.Referee_Stage = nil
	teamNames := map[string]int{}
	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Fatal(err)
		}

		if refereeMsg != nil {
			teamNames[*refereeMsg.Yellow.Name]++
			teamNames[*refereeMsg.Blue.Name]++

			if lastStage == nil || *refereeMsg.Stage != *lastStage {
				log.Printf("Found stage '%v'", referee.Referee_Stage_name[int32(*refereeMsg.Stage)])
			}

			if lastStage == nil {
				lastStage = new(referee.Referee_Stage)
			}
			*lastStage = *refereeMsg.Stage

			if logCutter.Stopped() &&
				*refereeMsg.Command != referee.Referee_HALT &&
				*refereeMsg.Stage != referee.Referee_POST_GAME {
				log.Println("Found non POST_GAME stage. Starting log file.")
				logCutter.Start()
			}

			if logCutter.Running() &&
				*refereeMsg.Command == referee.Referee_HALT &&
				*refereeMsg.Stage == referee.Referee_NORMAL_FIRST_HALF_PRE {
				log.Println("Found NORMAL_FIRST_HALF_PRE stage. Stopping log file.")
				logCutter.Stop()
			}

			logCutter.Update(refereeMsg)
			logCutter.Write(logMessage)

			if logCutter.Running() &&
				*refereeMsg.Command == referee.Referee_HALT &&
				*refereeMsg.Stage == referee.Referee_POST_GAME {
				log.Println("Found POST_GAME stage. Closing log file.")
				logCutter.Stop()
			}
		} else {
			logCutter.Write(logMessage)
		}
	}

	log.Printf("Teams: %v", teamNames)
}

func (l *LogCutter) Stop() {
	if l.logWriter == nil {
		return
	}

	log.Print("Stop log writer")

	if err := l.logWriter.Close(); err != nil {
		log.Fatal("Could not close log writer: ", err)
	}
	l.logWriter = nil

	if l.lastRefereeMsg == nil || l.firstRefereeMsg == nil {
		log.Println("No referee data found.")
	} else if *l.lastRefereeMsg.Stage == referee.Referee_NORMAL_FIRST_HALF_PRE {
		log.Println("Log ends with NORMAL_FIRST_HALF_PRE stage. Skipping.")
	} else {
		newLogFilename := logFileName(l.firstRefereeMsg, l.lastRefereeMsg)
		if err := shorten(newLogFilename, l.lastRefereeMsg); err != nil {
			log.Fatalf("Could not shorten file from '%v' to '%v'.", tmpLogFilename, newLogFilename)
		} else {
			log.Println("Saved to", newLogFilename)
		}
	}
	l.firstRefereeMsg = nil
	l.lastRefereeMsg = nil

	if err := os.Remove(tmpLogFilename); err != nil {
		log.Fatal("Could not remove tmp log file:", err)
	}
}

func shorten(newLogFilename string, lastRefereeMsg *referee.Referee) error {
	log.Printf("Shortening %v to %v", tmpLogFilename, newLogFilename)
	logReader, err := persistence.NewReader(tmpLogFilename)
	if err != nil {
		return err
	}

	logWriter, err := persistence.NewWriter(newLogFilename)
	if err != nil {
		return err
	}

	channel := logReader.CreateChannel()

	var lastRefereeMsgWithoutTimestamp referee.Referee

	proto.Merge(&lastRefereeMsgWithoutTimestamp, lastRefereeMsg)
	*lastRefereeMsgWithoutTimestamp.PacketTimestamp = 0

	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Fatal(err)
		}

		if err := logWriter.Write(logMessage); err != nil {
			log.Println("Could not write log message:", err)
		}

		if refereeMsg != nil {
			*refereeMsg.PacketTimestamp = 0
			if proto.Equal(refereeMsg, &lastRefereeMsgWithoutTimestamp) {
				break
			}
		}
	}

	if err := logWriter.Close(); err != nil {
		log.Fatal("Could not close log writer: ", err)
	}
	return nil
}

func getRefereeMsg(logMessage *persistence.Message) (refereeMsg *referee.Referee, err error) {
	if logMessage.MessageType.Id != persistence.MessageSslRefbox2013 {
		return
	}

	refereeMsg = new(referee.Referee)
	if err := proto.Unmarshal(logMessage.Message, refereeMsg); err != nil {
		err = errors.Wrap(err, "Could not parse referee message")
	}
	return
}

func logFileName(firstRefereeMsg, lastRefereeMsg *referee.Referee) string {
	teamNameYellow := strings.Replace(*lastRefereeMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*lastRefereeMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*firstRefereeMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	return fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
}
