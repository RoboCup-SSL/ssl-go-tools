package player

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

type Broadcaster struct {
	Slots                map[persistence.MessageId]*BroadcasterSlot
	reader               *persistence.Reader
	SkipNonRunningStages bool
}

type BroadcasterSlot struct {
	ReceivedMessages int
	MessageType      persistence.MessageType
	client           *sslnet.UdpClient
}

func NewBroadcaster() Broadcaster {
	return Broadcaster{Slots: make(map[persistence.MessageId]*BroadcasterSlot, 0)}
}

func (b *Broadcaster) AddSlot(messageType persistence.MessageType, address, nif string) {
	b.Slots[messageType.Id] = &BroadcasterSlot{client: sslnet.NewUdpClient(address, nif), MessageType: messageType}
}

func (b *Broadcaster) Start(filename string, startTimestamp int64) error {
	reader, err := persistence.NewReader(filename)
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
	curStage := gc.Referee_Stage(-1)
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

		if b.SkipNonRunningStages && msg.MessageType.Id == persistence.MessageSslRefbox2013 {
			var refMsg gc.Referee
			if err := proto.Unmarshal(msg.Message, &refMsg); err != nil {
				log.Println("Could not parse referee message", err)
			} else {
				curStage = *refMsg.Stage
			}
		}
	}
}
func isRunningStage(stage gc.Referee_Stage) bool {
	switch stage {
	case -1:
		return true
	case gc.Referee_NORMAL_FIRST_HALF:
		return true
	case gc.Referee_NORMAL_SECOND_HALF:
		return true
	case gc.Referee_EXTRA_FIRST_HALF:
		return true
	case gc.Referee_EXTRA_SECOND_HALF:
		return true
	}
	return false
}
