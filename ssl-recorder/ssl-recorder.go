package main

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Logger struct {
	logWriter             *sslproto.LogWriter
	refereeAddr           string
	visionAddr            string
	numberOfFramesVision  int
	numberOfFramesReferee int
	mutex                 sync.Mutex
}

const maxDatagramSize = 8192 * 2

func main() {

	logger := NewLogger()
	logger.Start()
	logger.registerToInterrupt()

	refListener := logger.openConnection(logger.refereeAddr)
	go logger.AcceptPackages(refListener, sslproto.MESSAGE_SSL_REFBOX_2013)

	visionListener := logger.openConnection(logger.visionAddr)
	go logger.AcceptPackages(visionListener, sslproto.MESSAGE_SSL_VISION_2014)

	for {
		fmt.Printf("\rCaptured %d vision frames and %d referee frames", logger.numberOfFramesVision, logger.numberOfFramesReferee)
		time.Sleep(time.Millisecond * 500)
	}
}

func NewLogger() Logger {
	return Logger{refereeAddr: "224.5.23.1:10003", visionAddr: "224.5.23.2:10006", numberOfFramesReferee: 0, numberOfFramesVision: 0}
}

func (l *Logger) Start() (err error) {
	err = l.openLogWriter()
	if err != nil {
		return
	}
	return
}

func (l *Logger) openLogWriter() (err error) {
	nowStr := time.Now().Format("2006-01-02_15-04-05")
	logFileName := "logs/" + nowStr + ".log"
	l.logWriter, err = sslproto.NewLogWriter(logFileName)
	if err != nil {
		log.Fatal("could not open log file for write: ", err)
	}
	return
}

func (l *Logger) Close() (err error) {
	err = l.logWriter.Close()
	if err != nil {
		return
	}
	return
}

func (l *Logger) registerToInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			l.Close()
			os.Exit(0)
		}
	}()
}

func (l *Logger) openConnection(address string) (listener *net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	listener, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("could not connect to ", address)
	}
	listener.SetReadBuffer(maxDatagramSize)
	log.Printf("Listening on %s", address)
	return
}

func (l *Logger) AcceptPackages(listener *net.UDPConn, messageType int) {
	for {
		data := make([]byte, maxDatagramSize)
		n, _, err := listener.ReadFromUDP(data)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		if err != nil {
			log.Print("Could not parse referee message: ", err)
		} else {
			timestamp := time.Now().UnixNano()
			logMessage := sslproto.LogMessage{Timestamp: timestamp, MessageType: messageType, Message: data[:n]}
			l.mutex.Lock()
			l.logWriter.WriteMessage(&logMessage)
			l.mutex.Unlock()
			if messageType == sslproto.MESSAGE_SSL_REFBOX_2013 {
				l.numberOfFramesReferee++
			} else if messageType == sslproto.MESSAGE_SSL_VISION_2014 {
				l.numberOfFramesVision++
			}
		}
	}
}
