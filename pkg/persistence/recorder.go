package persistence

import (
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

const maxDatagramSize = 8192 * 2

type Recorder struct {
	Slots  []*Slot
	writer Writer
	mutex  sync.Mutex
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
	return nil
}

func (r *Recorder) Stop() error {
	return r.writer.Close()
}

func (r *Recorder) RegisterToInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			err := r.Stop()
			if err != nil {
				log.Println("Could not stop recorder: ", err)
			}
			os.Exit(0)
		}
	}()
}

func (r *Recorder) openLogWriter() error {
	nowStr := time.Now().Format("2006-01-02_15-04-05")
	logFileName := nowStr + ".log.gz"
	writer, err := NewWriter(logFileName)
	if err != nil {
		return errors.Errorf("could not open log file for write: %v", err)
	}
	r.writer = writer
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
	for {
		data := make([]byte, maxDatagramSize)
		n, _, err := listener.ReadFromUDP(data)
		if err != nil {
			log.Print("ReadFromUDP failed:", err)
			return
		}

		if err != nil {
			log.Print("Could not parse message: ", err)
		} else {
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
}
