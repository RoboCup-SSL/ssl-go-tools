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

	tmpLogFilename := "tmp.log.gz"
	logWriter, err := persistence.NewWriter(tmpLogFilename)
	if err != nil {
		log.Println("Could not open temporary writer:", err)
		return
	}

	var lastRefereeMsg *sslproto.SSL_Referee = nil
	var lastStage *sslproto.SSL_Referee_Stage = nil
	skipped := 0
	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Fatal(err)
		}

		if refereeMsg != nil && lastRefereeMsg != nil && *refereeMsg.CommandCounter < *lastRefereeMsg.CommandCounter {
			skipped++
			continue
		}

		if refereeMsg != nil && (lastStage == nil || *refereeMsg.Stage != *lastStage) {
			switch *refereeMsg.Stage {
			case sslproto.SSL_Referee_NORMAL_FIRST_HALF:
				log.Println("Found first half")
			case sslproto.SSL_Referee_NORMAL_SECOND_HALF:
				log.Println("Found second half")
			case sslproto.SSL_Referee_EXTRA_FIRST_HALF:
				log.Println("Found extra first half")
			case sslproto.SSL_Referee_EXTRA_SECOND_HALF:
				log.Println("Found extra second half")
			case sslproto.SSL_Referee_PENALTY_SHOOTOUT:
				log.Println("Found shootout")
			case sslproto.SSL_Referee_POST_GAME:
				log.Println("Found post game")
			}
		}

		if refereeMsg != nil {
			lastRefereeMsg = refereeMsg

			if lastStage == nil {
				lastStage = new(sslproto.SSL_Referee_Stage)
			}
			*lastStage = *refereeMsg.Stage

			if *refereeMsg.Stage == sslproto.SSL_Referee_POST_GAME {
				continue
			}
		}

		if err := logWriter.Write(logMessage); err != nil {
			log.Println("Could not write log message:", err)
		}
	}

	if err := logWriter.Close(); err != nil {
		log.Println("Could not close log writer: ", err)
	}

	if lastRefereeMsg == nil {
		if err := os.Remove(tmpLogFilename); err != nil {
			log.Println("Could not remove tmp log file:", err)
		}
		return
	}

	if skipped > 0 {
		log.Printf("Skipped %d referee messages, because they were out of order (probably a second referee source).", skipped)
	}

	newLogFilename := logFileName(lastRefereeMsg)
	if err := os.Rename(tmpLogFilename, newLogFilename); err != nil {
		log.Printf("Could not rename file from '%v' to '%v'.", tmpLogFilename, newLogFilename)
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
