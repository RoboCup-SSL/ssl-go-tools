package persistence

import (
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"time"
)

const maxDatagramSize = 8192 * 2

type Recorder struct {
	Slots   []*Slot
	writer  Writer
	running bool
	mutex   sync.Mutex
}

type Slot struct {
	ReceivedMessages int
	MessageType      MessageType
	address          string
}

func NewRecorder() Recorder {
	return Recorder{Slots: make([]*Slot, 0)}
}

func (r *Recorder) AddSlot(messageType MessageType, address string) {
	r.Slots = append(r.Slots, &Slot{address: address, MessageType: messageType})
}

func (r *Recorder) Start() error {
	if r.running {
		return errors.New("Recorder already started")
	}
	if err := r.openLogWriter(); err != nil {
		return err
	}
	for name, slot := range r.Slots {
		listener, err := openConnection(slot.address)
		if err != nil {
			log.Printf("Could not open connection for %v on %v", name, slot.address)
		} else {
			go r.acceptMessages(listener, slot)
		}
	}
	r.running = true
	return nil
}

func (r *Recorder) Stop() error {
	if !r.running {
		return nil
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

func openConnection(address string) (listener *net.UDPConn, err error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}
	listener, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return
	}
	err = listener.SetReadBuffer(maxDatagramSize)
	if err != nil {
		return
	}
	log.Printf("Listening on %s", address)
	return
}

func (r *Recorder) acceptMessages(listener *net.UDPConn, slot *Slot) {
	data := make([]byte, maxDatagramSize)
	for r.running {
		n, _, err := listener.ReadFromUDP(data)
		if err != nil {
			log.Print("ReadFromUDP failed:", err)
			return
		}

		timestamp := time.Now().UnixNano()
		logMessage := Message{Timestamp: timestamp, MessageType: slot.MessageType, Message: data[:n]}
		r.mutex.Lock()
		err = r.writer.Write(&logMessage)
		if err != nil {
			log.Println("Could not write log message: ", err)
		}
		r.mutex.Unlock()
		slot.ReceivedMessages++
	}
}
