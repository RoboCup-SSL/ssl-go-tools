package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

// NewBroadcaster creates a new UDP multicast connection on which to broadcast
func NewBroadcaster(address string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	conn.SetReadBuffer(maxDatagramSize)
	log.Println("Connected to", address)

	return conn

}

func main() {
	logFile := flag.String("logfile", "", "The log file to play")
	skipNonRunningStages := flag.Bool("skip", false, "Skip frames while not in a running stage")

	flag.Parse()

	if logFile == nil {
		log.Fatalln("Missing logfile")
	}

	logReader, err := sslproto.NewLogReader(*logFile)
	if err != nil {
		return
	}
	defer logReader.Close()

	legacyVisionConn := NewBroadcaster("224.5.23.2:10005")
	visionConn := NewBroadcaster("224.5.23.2:10006")
	refereeConn := NewBroadcaster("224.5.23.1:10003")

	startTime := time.Now()
	refTimestamp := int64(0)
	curStage := sslproto.SSL_Referee_Stage(-1)
	for logReader.HasMessage() {
		msg, err := logReader.ReadMessage()
		if err != nil {
			log.Println("Could not read message", err)
			continue
		}
		if isRunningStage(curStage) {

			if refTimestamp != 0 {
				realElapsed := time.Now().Sub(startTime).Nanoseconds()
				msgElapsed := msg.Timestamp - refTimestamp
				sleepTime := msgElapsed - realElapsed
				time.Sleep(time.Duration(sleepTime))
			} else {
				startTime = time.Now()
				refTimestamp = msg.Timestamp
			}

			switch msg.MessageType {
			case sslproto.MESSAGE_SSL_VISION_2010:
				legacyVisionConn.Write(msg.Message)
			case sslproto.MESSAGE_SSL_VISION_2014:
				visionConn.Write(msg.Message)
			case sslproto.MESSAGE_SSL_REFBOX_2013:
				refereeConn.Write(msg.Message)
			default:
				log.Println("Unknown message type: ", msg.MessageType)
			}
		} else {
			refTimestamp = 0
		}

		if *skipNonRunningStages && msg.MessageType == sslproto.MESSAGE_SSL_REFBOX_2013 {
			refMsg, err := msg.ParseReferee()
			if err != nil {
				log.Println("Could not parse referee message:", err)
			} else {
				curStage = *refMsg.Stage
			}
		}
	}
}
func isRunningStage(stage sslproto.SSL_Referee_Stage) bool {
	switch stage {
	case -1:
		return true
	case sslproto.SSL_Referee_NORMAL_FIRST_HALF:
		return true
	case sslproto.SSL_Referee_NORMAL_SECOND_HALF:
		return true
	case sslproto.SSL_Referee_EXTRA_FIRST_HALF:
		return true
	case sslproto.SSL_Referee_EXTRA_SECOND_HALF:
		return true
	}
	return false
}
