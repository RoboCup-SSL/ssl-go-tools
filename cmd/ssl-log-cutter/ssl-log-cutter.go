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
var logWriter *persistence.Writer = nil
var firstRefereeMsg *referee.Referee = nil
var lastRefereeMsg *referee.Referee = nil

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

	if logWriter != nil {
		closeLogWriter(logWriter)
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
				firstRefereeMsg = refereeMsg
				logWriter, err = persistence.NewWriter(tmpLogFilename)
				if err != nil {
					log.Fatal("Could not open temporary writer:", err)
				}
			}

			if logWriter != nil && lastRefereeMsg != nil &&
				*refereeMsg.Command == referee.Referee_HALT &&
				(*refereeMsg.Stage == referee.Referee_POST_GAME ||
					*refereeMsg.Stage == referee.Referee_NORMAL_FIRST_HALF_PRE) {

				if *refereeMsg.Stage == referee.Referee_POST_GAME {
					if err := logWriter.Write(logMessage); err != nil {
						log.Println("Could not write log message:", err)
					}
				}

				log.Print("Stop log writer")
				closeLogWriter(logWriter)
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

	log.Printf("Found %d valid referee messages, skipped %d unreasonable referee messages with these team names: %v",
		numRefereeMessages, numSkippedRefereeMessages, unreasonableTeamNames)
}

func closeLogWriter(logWriter *persistence.Writer) {
	if err := logWriter.Close(); err != nil {
		log.Fatal("Could not close log writer: ", err)
	}
	if lastRefereeMsg == nil || firstRefereeMsg == nil || *lastRefereeMsg.Stage == referee.Referee_NORMAL_FIRST_HALF_PRE {
		log.Println("No reasonable referee data found.")
	} else {
		newLogFilename := logFileName()
		if err := shorten(newLogFilename); err != nil {
			log.Fatalf("Could not shorten file from '%v' to '%v'.", tmpLogFilename, newLogFilename)
		} else {
			log.Println("Saved to", newLogFilename)
		}
	}
	firstRefereeMsg = nil
	lastRefereeMsg = nil

	if err := os.Remove(tmpLogFilename); err != nil {
		log.Fatal("Could not remove tmp log file:", err)
	}
}

func shorten(newLogFilename string) error {
	logReader, err := persistence.NewReader(tmpLogFilename)
	if err != nil {
		return err
	}

	logWriter, err = persistence.NewWriter(newLogFilename)
	if err != nil {
		return err
	}

	channel := logReader.CreateChannel()

	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Fatal(err)
		}

		if err := logWriter.Write(logMessage); err != nil {
			log.Println("Could not write log message:", err)
		}
		if refereeMsg != nil && *refereeMsg.CommandCounter == *lastRefereeMsg.CommandCounter {
			break
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

func logFileName() string {
	teamNameYellow := strings.Replace(*lastRefereeMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*lastRefereeMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*firstRefereeMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	return fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
}
