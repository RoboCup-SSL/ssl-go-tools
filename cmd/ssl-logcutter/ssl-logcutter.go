package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"log"
	"strings"
	"time"
)

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatalln("Pass one or more log files")
	}

	for _, arg := range args {
		log.Println("Processing", arg)
		process(arg)
	}
}

func process(filename string) {
	logReader, err := persistence.NewReader(filename)
	if err != nil {
		log.Println("Could not process log file:", err)
		return
	}
	defer logReader.Close()

	channel := logReader.CreateChannel()

	var logWriter *persistence.Writer

	for logMessage := range channel {
		if logMessage.MessageType.Id == persistence.MessageSslRefbox2013 {
			refereeMsg, err := logMessage.ParseReferee()
			if err != nil {
				log.Println("Could not parse referee message. Stop processing.", err)
				return
			}
			switch *refereeMsg.Stage {
			case sslproto.SSL_Referee_NORMAL_FIRST_HALF,
				sslproto.SSL_Referee_NORMAL_SECOND_HALF,
				sslproto.SSL_Referee_EXTRA_FIRST_HALF,
				sslproto.SSL_Referee_EXTRA_SECOND_HALF:
				// we have to start at a point, where team names are guarantied to have been entered
				// the NORMAL_START command is only allowed, if team names were entered. So we will begin at least there
				// we are not that much interested in the kick-off preparation, so we start with the transition to the half-stages
				if logWriter == nil {
					logFileName := logFileName(refereeMsg, logMessage)
					w, err := persistence.NewWriter(logFileName)
					if err != nil {
						log.Println("Can not create log writer: ", err)
						return
					}
					logWriter = &w
					log.Println("Saving to", logFileName)
				}
			case sslproto.SSL_Referee_POST_GAME:
				log.Println("POST_GAME found. Stop processing.")
				return
			}
		}
		if logWriter != nil {
			logWriter.Write(logMessage)
		}
	}
	log.Println("Processing done")
}

func logFileName(refereeMsg *sslproto.SSL_Referee, r *persistence.Message) string {
	teamNameYellow := strings.Replace(*refereeMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refereeMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, r.Timestamp).Format("2006-01-02_15-04")
	logFileName := fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
	return logFileName
}
