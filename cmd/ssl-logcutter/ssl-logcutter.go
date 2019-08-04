package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/pkg/errors"
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

	var lastRefereeMsg *sslproto.SSL_Referee = nil
	var lastStage *sslproto.SSL_Referee_Stage = nil
	numSkippedRefereeMessages := 0
	numRefereeMessages := 0
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
				continue
			}

			if lastStage == nil || *refereeMsg.Stage != *lastStage {
				log.Printf("Found stage '%v'", sslproto.SSL_Referee_Stage_name[int32(*refereeMsg.Stage)])
			}

			if logWriter == nil &&
				*refereeMsg.Stage != sslproto.SSL_Referee_POST_GAME &&
				*refereeMsg.Command != sslproto.SSL_Referee_HALT {
				log.Print("Start log writer")
				logWriter, err = persistence.NewWriter(tmpLogFilename)
				if err != nil {
					log.Fatal("Could not open temporary writer:", err)
				}
			}

			if logWriter != nil &&
				*refereeMsg.Stage == sslproto.SSL_Referee_POST_GAME {
				log.Print("Stop log writer")
				closeLogWriter(logWriter, lastRefereeMsg)
				logWriter = nil
			}

			if lastStage == nil {
				lastStage = new(sslproto.SSL_Referee_Stage)
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

	log.Printf("Found %d valid referee messages, skipped %d unreasonable referee messages.",
		numRefereeMessages, numSkippedRefereeMessages)
}

func closeLogWriter(logWriter *persistence.Writer, lastRefereeMsg *sslproto.SSL_Referee) {
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

func getRefereeMsg(logMessage *persistence.Message) (refereeMsg *sslproto.SSL_Referee, err error) {
	if logMessage.MessageType.Id != persistence.MessageSslRefbox2013 {
		return
	}
	refereeMsg, err = logMessage.ParseReferee()
	if err != nil {
		err = errors.Wrap(err, "Could not parse referee message")
	}
	return
}

func logFileName(refereeMsg *sslproto.SSL_Referee) string {
	teamNameYellow := strings.Replace(*refereeMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refereeMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*refereeMsg.PacketTimestamp*1000)).Format("2006-01-02_15-04")
	logFileName := fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
	return logFileName
}
