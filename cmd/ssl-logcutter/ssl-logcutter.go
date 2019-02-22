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

	var logWriter persistence.Writer

	for logMessage := range channel {
		refereeMsg, err := getRefereeMsg(logMessage)
		if err != nil {
			log.Println(err)
			break
		}
		if refereeMsg != nil && !logWriter.Open {
			logWriter, err = createLogWriter(refereeMsg)
		}
		if refereeMsg != nil && *refereeMsg.Stage == sslproto.SSL_Referee_POST_GAME {
			log.Println("POST_GAME found. Stop processing.")
			break
		}
		if logWriter.Open {
			err = logWriter.Write(logMessage)
		}
	}

	err = logWriter.Close()
	if err != nil {
		log.Println("Could not close log writer: ", err)
	}
	err = logReader.Close()
	if err != nil {
		log.Println("Could not close log reader: ", err)
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

func createLogWriter(refereeMsg *sslproto.SSL_Referee) (logWriter persistence.Writer, err error) {
	switch *refereeMsg.Stage {
	case sslproto.SSL_Referee_NORMAL_FIRST_HALF,
		sslproto.SSL_Referee_NORMAL_SECOND_HALF,
		sslproto.SSL_Referee_EXTRA_FIRST_HALF,
		sslproto.SSL_Referee_EXTRA_SECOND_HALF:
		// we have to start at a point, where team names are guarantied to have been entered
		// the NORMAL_START command is only allowed, if team names were entered. So we will begin at least there
		// we are not that much interested in the kick-off preparation, so we start with the transition to the half-stages
		logFileName := logFileName(refereeMsg)
		logWriter, err = persistence.NewWriter(logFileName)
		if err != nil {
			err = errors.Wrap(err, "Can not create log writer")
		}
		log.Println("Saving to", logFileName)
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
