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
	if len(args) != 1 {
		log.Fatal("Pass a log file to analyze")
	}
	fileName := flag.Args()[0]
	reader, err := sslproto.NewLogReader(fileName)
	if err != nil {
		log.Fatal("Could not open log file: ", fileName)
	}

	c := make(chan *sslproto.LogMessage)
	go reader.CreateLogMessageChannel(c)

	format := "%20v %25s %25s\n"
	fmt.Printf(format, "Receive Time", "Referee", "AutoRef")
	for m := range c {
		t := time.Unix(0, m.Timestamp).Format("2006-01-02 15:04:05")
		switch m.MessageType {
		case sslproto.MESSAGE_SSL_REFBOX_2013:
			ref := m.ParseReferee()
			fmt.Printf(format, t, ref.Command, "")
		case sslproto.MESSAGE_SSL_REFBOX_RCON_2018:
			rcon := m.ParseRefereeRemoteControlRequest()
			fmt.Printf(format, t, "", rcon.Command)
		}
	}
}
