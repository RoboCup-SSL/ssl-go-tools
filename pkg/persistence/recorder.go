package persistence

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"github.com/pkg/errors"
)

type Recorder struct {
	Slots            []*RecorderSlot
	writer           Writer
	recording        bool
	paused           bool
	receiving        bool
	mutex            sync.Mutex
	messageConsumers []func(*Message, *net.UDPAddr)
}

type RecorderSlot struct {
	ReceivedMessages int
	MessageType      MessageType
	server           *sslnet.MulticastServer
}

func NewRecorder() Recorder {
	return Recorder{Slots: make([]*RecorderSlot, 0)}
}

func (r *Recorder) IsRecording() bool {
	return r.recording
}

func (r *Recorder) IsPaused() bool {
	return r.paused
}

func (r *Recorder) SetPaused(paused bool) {
	r.paused = paused
}

func (r *Recorder) AddSlot(messageType MessageType, address string) {
	r.Slots = append(r.Slots, &RecorderSlot{
		MessageType: messageType,
		server:      sslnet.NewMulticastServer(address),
	})
}

func (r *Recorder) AddMessageConsumer(consumer func(*Message, *net.UDPAddr)) {
	r.messageConsumers = append(r.messageConsumers, consumer)
}

func (r *Recorder) StartReceiving() {
	if r.receiving {
		return
	}
	for _, slot := range r.Slots {
		slot.server.Consumer = r.slotConsumer(slot)
		slot.server.Start()
	}
	r.receiving = true
}

func (r *Recorder) StopReceiving() {
	if !r.receiving {
		return
	}
	for _, slot := range r.Slots {
		slot.server.Stop()
	}
	r.receiving = false
}

func (r *Recorder) StartRecording(logFileName string) error {
	if r.recording {
		return errors.New("Recorder already started")
	}
	if err := r.openLogWriter(logFileName); err != nil {
		return err
	}
	r.paused = false
	r.recording = true
	return nil
}

func (r *Recorder) StopRecording() error {
	if !r.recording {
		return nil
	}
	r.recording = false
	return r.writer.Close()
}

func (r *Recorder) slotConsumer(slot *RecorderSlot) func(bytes []byte, addr *net.UDPAddr) {
	return func(bytes []byte, addr *net.UDPAddr) {
		r.processSlotMessage(slot, bytes, addr)
	}
}

func (r *Recorder) openLogWriter(logFileName string) error {
	writer, err := NewWriter(logFileName)
	if err != nil {
		return errors.Errorf("could not open log file for write: %v", err)
	}
	r.writer = *writer
	return nil
}

func (r *Recorder) processSlotMessage(slot *RecorderSlot, data []byte, addr *net.UDPAddr) {
	timestamp := time.Now().UnixNano()
	logMessage := Message{Timestamp: timestamp, MessageType: slot.MessageType, Message: data}
	r.mutex.Lock()

	if r.recording && !r.paused {
		if err := r.writer.Write(&logMessage); err != nil {
			log.Println("Could not write log message: ", err)
		}
		slot.ReceivedMessages++
	}

	for _, consumer := range r.messageConsumers {
		consumer(&logMessage, addr)
	}
	r.mutex.Unlock()
}
