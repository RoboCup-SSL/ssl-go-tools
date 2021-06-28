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

var tmpLogFilename = "tmp.log.gz"

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

	var logWriter *persistence.Writer = nil

	var lastRefereeMsg *referee.Referee = nil
	var lastStage *referee.Referee_Stage = nil
	numSkippedRefereeMessages := 0
	numRefereeMessages := 0
	unreasonableTeamNames := map[string]int{}
	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Fatal(err)
		}

		if refereeMsg != nil {
			numRefereeMessages++

			if *refereeMsg.Yellow.Name == "" ||
				*refereeMsg.Blue.Name == "" ||
				*refereeMsg.Yellow.Name == *refereeMsg.Blue.Name {
				numSkippedRefereeMessages++
				if *refereeMsg.Yellow.Name != "" && *refereeMsg.Yellow.Name == *refereeMsg.Blue.Name {
					i, ok := unreasonableTeamNames[*refereeMsg.Yellow.Name]
					if ok {
						unreasonableTeamNames[*refereeMsg.Yellow.Name] = i + 1
					} else {
						unreasonableTeamNames[*refereeMsg.Yellow.Name] = 1
					}
				}
				continue
			}

			if lastStage == nil || *refereeMsg.Stage != *lastStage {
				log.Printf("Found stage '%v'", referee.Referee_Stage_name[int32(*refereeMsg.Stage)])
			}

			if logWriter == nil &&
				*refereeMsg.Stage != referee.Referee_POST_GAME &&
				*refereeMsg.Command != referee.Referee_HALT {
				log.Print("Start log writer")
				logWriter, err = persistence.NewWriter(tmpLogFilename)
				if err != nil {
					log.Fatal("Could not open temporary writer:", err)
				}
			}

			if logWriter != nil && lastRefereeMsg != nil &&
				*refereeMsg.Command == referee.Referee_HALT &&
				(*refereeMsg.Stage == referee.Referee_POST_GAME ||
					*refereeMsg.Stage == referee.Referee_NORMAL_FIRST_HALF_PRE) {
				log.Print("Stop log writer")
				closeLogWriter(logWriter, lastRefereeMsg)
				logWriter = nil
			}

			if lastStage == nil {
				lastStage = new(referee.Referee_Stage)
			}
			*lastStage = *refereeMsg.Stage

			lastRefereeMsg = refereeMsg
		}

		if logWriter != nil {
			if err := logWriter.Write(logMessage); err != nil {
				log.Println("Could not write log message:", err)
			}
		}
	}

	if logWriter != nil {
		closeLogWriter(logWriter, lastRefereeMsg)
	}

	log.Printf("Found %d valid referee messages, skipped %d unreasonable referee messages with these team names: %v",
		numRefereeMessages, numSkippedRefereeMessages, unreasonableTeamNames)
}

func closeLogWriter(logWriter *persistence.Writer, lastRefereeMsg *referee.Referee) {
	if err := logWriter.Close(); err != nil {
		log.Fatal("Could not close log writer: ", err)
	}
	if lastRefereeMsg == nil {
		if err := os.Remove(tmpLogFilename); err != nil {
			log.Fatal("Could not remove tmp log file:", err)
		}
		return
	}
	newLogFilename := logFileName(lastRefereeMsg)
	if err := os.Rename(tmpLogFilename, newLogFilename); err != nil {
		log.Fatalf("Could not rename file from '%v' to '%v'.", tmpLogFilename, newLogFilename)
	} else {
		log.Println("Saved to", newLogFilename)
	}
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

func logFileName(refereeMsg *referee.Referee) string {
	teamNameYellow := strings.Replace(*refereeMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refereeMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*refereeMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	logFileName := fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
	return logFileName
}
