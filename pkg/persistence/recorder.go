package persistence

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"time"
)

type Recorder struct {
	Slots   []*RecorderSlot
	writer  Writer
	running bool
	mutex   sync.Mutex
}

type RecorderSlot struct {
	ReceivedMessages int
	MessageType      MessageType
	server           *sslnet.MulticastServer
}

func NewRecorder() Recorder {
	return Recorder{Slots: make([]*RecorderSlot, 0)}
}

func (r *Recorder) AddSlot(messageType MessageType, address string) {
	r.Slots = append(r.Slots, &RecorderSlot{
		MessageType: messageType,
		server:      sslnet.NewMulticastServer(address),
	})
}

func (r *Recorder) Start() error {
	if r.running {
		return errors.New("Recorder already started")
	}
	if err := r.openLogWriter(); err != nil {
		return err
	}
	for _, slot := range r.Slots {
		slot.server.Consumer = r.slotConsumer(slot)
		slot.server.Start()
	}
	r.running = true
	return nil
}

func (r *Recorder) slotConsumer(slot *RecorderSlot) func(bytes []byte, addr *net.UDPAddr) {
	return func(bytes []byte, addr *net.UDPAddr) {
		r.processSlotMessage(slot, bytes)
	}
}

func (r *Recorder) Stop() error {
	if !r.running {
		return nil
	}
	for _, slot := range r.Slots {
		slot.server.Stop()
	}
	r.running = false
	return r.writer.Close()
}

func (r *Recorder) IsRunning() bool {
	return r.running
}

func (r *Recorder) openLogWriter() error {
	nowStr := time.Now().Format("2006-01-02_15-04-05")
	logFileName := nowStr + ".log.gz"
	writer, err := NewWriter(logFileName)
	if err != nil {
		return errors.Errorf("could not open log file for write: %v", err)
	}
	r.writer = *writer
	return nil
}

func (r *Recorder) processSlotMessage(slot *RecorderSlot, data []byte) {
	timestamp := time.Now().UnixNano()
	logMessage := Message{Timestamp: timestamp, MessageType: slot.MessageType, Message: data}
	r.mutex.Lock()
	if err := r.writer.Write(&logMessage); err != nil {
		log.Println("Could not write log message: ", err)
	}
	r.mutex.Unlock()
	slot.ReceivedMessages++
}
