package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"log"
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

func process(logFile string) {
	logReader, err := sslproto.NewLogReader(logFile)
	if err != nil {
		log.Println("Could not process log file:", err)
		return
	}
	defer logReader.Close()

	channel := make(chan *sslproto.LogMessage, 100)
	go logReader.CreateLogMessageChannel(channel)

	var logWriter *sslproto.LogWriter

	for logMessage := range channel {
		if logMessage.MessageType == sslproto.MESSAGE_SSL_REFBOX_2013 {
			refereeMsg := logMessage.ParseReferee()
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
					logWriter, err = sslproto.NewLogWriter(logFileName)
					if err != nil {
						log.Println("Can not create log writer: ", err)
						return
					}
					log.Println("Saving to", logFileName)
				}
			case sslproto.SSL_Referee_POST_GAME:
				log.Println("POST_GAME found. Stop processing.")
				return
			}
		}
		if logWriter != nil {
			logWriter.WriteMessage(logMessage)
		}
	}
	log.Println("Processing done")
}

func logFileName(refereeMsg *sslproto.SSL_Referee, r *sslproto.LogMessage) string {
	teamNameYellow := *refereeMsg.Yellow.Name
	teamNameBlue := *refereeMsg.Blue.Name
	date := time.Unix(0, r.Timestamp).Format("2006-01-02_15-04")
	logFileName := fmt.Sprintf("%s_%s-vs-%s.log.gz", date, teamNameYellow, teamNameBlue)
	return logFileName
}
