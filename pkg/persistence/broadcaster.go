package persistence

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"log"
	"net"
	"time"
)

type Broadcaster struct {
	Slots                []*Slot
	conns                map[MessageId]*net.UDPConn
	reader               *Reader
	SkipNonRunningStages bool
}

func NewBroadcaster() Broadcaster {
	return Broadcaster{Slots: make([]*Slot, 0), conns: make(map[MessageId]*net.UDPConn)}
}

// NewBroadcaster creates a new UDP multicast connection on which to broadcast
func connectForPublishing(address string) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	err = conn.SetReadBuffer(maxDatagramSize)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Connected to", address)

	return conn
}

func (b *Broadcaster) AddSlot(messageType MessageType, address string) {
	b.Slots = append(b.Slots, &Slot{address: address, MessageType: messageType})
}

func (b *Broadcaster) Start(filename string, startTimestamp int64) error {
	reader, err := NewReader(filename)
	if err != nil {
		return err
	}
	b.reader = reader

	for _, slot := range b.Slots {
		conn := connectForPublishing(slot.address)
		b.conns[slot.MessageType.Id] = conn
	}

	b.publish(startTimestamp)
	return nil
}

func (b *Broadcaster) Stop() {
	for _, conn := range b.conns {
		err := conn.Close()
		if err != nil {
			fmt.Println("Could not close connection: ", err)
		}
	}
	if b.reader != nil {
		err := b.reader.Close()
		if err != nil {
			fmt.Println("Could not close reader: ", err)
		}
	}
}

func (b *Broadcaster) publish(startTimestamp int64) {
	startTime := time.Now()
	refTimestamp := int64(0)
	curStage := sslproto.Referee_Stage(-1)
	for b.reader.HasMessage() {
		msg, err := b.reader.ReadMessage()
		if err != nil {
			log.Fatal("Could not read message: ", err)
		}
		if msg.Timestamp < startTimestamp {
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

			if conn := b.conns[msg.MessageType.Id]; conn != nil {
				_, err := conn.Write(msg.Message)
				if err != nil {
					log.Println("Could not write message: ", err)
				}
			}
		} else {
			refTimestamp = 0
		}

		if b.SkipNonRunningStages && msg.MessageType.Id == MessageSslRefbox2013 {
			refMsg, err := msg.ParseReferee()
			if err != nil {
				log.Println("Could not parse referee message:", err)
			} else {
				curStage = *refMsg.Stage
			}
		}
	}
}
func isRunningStage(stage sslproto.Referee_Stage) bool {
	switch stage {
	case -1:
		return true
	case sslproto.Referee_NORMAL_FIRST_HALF:
		return true
	case sslproto.Referee_NORMAL_SECOND_HALF:
		return true
	case sslproto.Referee_EXTRA_FIRST_HALF:
		return true
	case sslproto.Referee_EXTRA_SECOND_HALF:
		return true
	}
	return false
}
