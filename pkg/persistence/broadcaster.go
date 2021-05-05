package persistence

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/referee"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"log"
	"time"
)

type Broadcaster struct {
	Slots                map[MessageId]*BroadcasterSlot
	reader               *Reader
	SkipNonRunningStages bool
}

type BroadcasterSlot struct {
	ReceivedMessages int
	MessageType      MessageType
	client           *sslnet.UdpClient
}

func NewBroadcaster() Broadcaster {
	return Broadcaster{Slots: make(map[MessageId]*BroadcasterSlot, 0)}
}

func (b *Broadcaster) AddSlot(messageType MessageType, address string) {
	b.Slots[messageType.Id] = &BroadcasterSlot{client: sslnet.NewUdpClient(address), MessageType: messageType}
}

func (b *Broadcaster) Start(filename string, startTimestamp int64) error {
	reader, err := NewReader(filename)
	if err != nil {
		return err
	}
	b.reader = reader

	for _, slot := range b.Slots {
		slot.client.Start()
	}

	b.publish(startTimestamp)
	return nil
}

func (b *Broadcaster) Stop() {
	for _, slot := range b.Slots {
		slot.client.Stop()
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
	curStage := referee.Referee_Stage(-1)
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

			if slot, ok := b.Slots[msg.MessageType.Id]; ok {
				slot.client.Send(msg.Message)
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
func isRunningStage(stage referee.Referee_Stage) bool {
	switch stage {
	case -1:
		return true
	case referee.Referee_NORMAL_FIRST_HALF:
		return true
	case referee.Referee_NORMAL_SECOND_HALF:
		return true
	case referee.Referee_EXTRA_FIRST_HALF:
		return true
	case referee.Referee_EXTRA_SECOND_HALF:
		return true
	}
	return false
}
