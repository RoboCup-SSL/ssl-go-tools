package main

import (
	"github.com/RoboCup-SSL/ssl-go-tools/protoconn"
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type Logger struct {
	numClients      int
	logWriter       *sslproto.LogWriter
	rconAddr        string
	rconListener    net.Listener
	refereeAddr     string
	refereeListener *net.UDPConn
}

const maxDatagramSize = 8192

func main() {

	logger := NewLogger()
	logger.Start()
	logger.registerToInterrupt()
	go logger.AcceptRefereePackages()
	logger.AcceptRemoteConnections()

}

func NewLogger() Logger {
	return Logger{numClients: 0, rconAddr: ":10007", refereeAddr: "224.5.23.1:10003"}
}

func (l *Logger) Start() (err error) {
	err = l.openLogWriter()
	if err != nil {
		return
	}
	err = l.openRconConnection()
	if err != nil {
		return
	}
	err = l.openRefereeConnection()
	if err != nil {
		return
	}
	return
}

func (l *Logger) openLogWriter() (err error) {
	nowStr := time.Now().Format("2006-01-02_15-04-05")
	logFileName := "log/" + nowStr + ".log"
	l.logWriter, err = sslproto.NewLogWriter(logFileName)
	if err != nil {
		log.Fatal("could not open log file for write: ", err)
	}
	return
}

func (l *Logger) openRconConnection() (err error) {
	l.rconListener, err = net.Listen("tcp", l.rconAddr)
	if err != nil {
		log.Fatal("could not connect to ", l.rconAddr)
	}
	log.Printf("Listening on %s", l.rconAddr)
	return
}

func (l *Logger) Close() (err error) {
	err = l.logWriter.Close()
	if err != nil {
		return
	}
	err = l.rconListener.Close()
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

func (l *Logger) AcceptRemoteConnections() {
	for {
		if conn, err := l.rconListener.Accept(); err == nil {
			go l.handleClientConnection(conn)
		}
	}
}

func (l *Logger) openRefereeConnection() (err error) {
	addr, err := net.ResolveUDPAddr("udp", l.refereeAddr)
	if err != nil {
		log.Fatal(err)
	}
	l.refereeListener, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("could not connect to ", l.refereeAddr)
	}
	l.refereeListener.SetReadBuffer(maxDatagramSize)
	log.Printf("Listening on %s", l.refereeAddr)
	return
}

func (l *Logger) AcceptRefereePackages() {
	lastCommandId := uint32(100000000)
	for {
		data := make([]byte, maxDatagramSize)
		n, _, err := l.refereeListener.ReadFromUDP(data)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		message, err := l.parseRefereeMessage(data[:n])
		if err != nil {
			log.Print("Could not parse referee message: ", err)
		} else if *message.CommandCounter != lastCommandId {
			log.Println("Received referee message:", message)
			timestamp := time.Now().UnixNano()
			messageType := sslproto.MESSAGE_SSL_REFBOX_2013
			logMessage := sslproto.LogMessage{Timestamp: timestamp, MessageType: messageType, Message: data[:n]}
			l.logWriter.WriteMessage(&logMessage)
			lastCommandId = *message.CommandCounter
		}
	}
}

func (l *Logger) parseRefereeMessage(data []byte) (message *sslproto.SSL_Referee, err error) {
	message = new(sslproto.SSL_Referee)
	if err = proto.Unmarshal(data, message); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal data")
	}
	return
}

func (l *Logger) handleClientConnection(clientConn net.Conn) {
	l.numClients++
	log.Printf("Connection established: %v, now %d clients", clientConn.RemoteAddr(), l.numClients)

	// Close the connection when the function exits
	defer clientConn.Close()

	for {
		err := l.handleClientRequest(clientConn)
		if errors.Cause(err) == io.EOF {
			// connection is closed
			break
		}
		if err != nil {
			log.Println("unable to handle client request: ", err)
		}
	}

	l.numClients--
	log.Printf("Connection closed: %v, now %d clients", clientConn.RemoteAddr(), l.numClients)
}

func (l *Logger) handleClientRequest(clientConnection net.Conn) error {

	request := new(sslproto.SSL_RefereeRemoteControlRequest)
	if err := protoconn.ReceiveMessage(clientConnection, request); err != nil {
		return errors.Wrap(err, "unable to receive request from client")
	}

	log.Println("Received rcon message:", request)
	timestamp := time.Now().UnixNano()
	messageType := sslproto.MESSAGE_SSL_REFBOX_RCON_2018
	data, err := proto.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "marshaling error")
	}
	logMessage := sslproto.LogMessage{Timestamp: timestamp, MessageType: messageType, Message: data}
	l.logWriter.WriteMessage(&logMessage)

	outcome := sslproto.SSL_RefereeRemoteControlReply_OK
	reply := &sslproto.SSL_RefereeRemoteControlReply{
		MessageId: request.MessageId,
		Outcome:   &outcome,
	}

	if err := protoconn.SendMessage(clientConnection, reply); err != nil {
		return errors.Wrap(err, "unable to send reply to client")
	}
	return nil
}
