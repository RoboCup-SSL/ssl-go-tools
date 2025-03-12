package sslnet

import (
	"log"
	"net"
	"sync"
	"time"
)

const maxDatagramSize = 8192

type MulticastServer struct {
	multicastAddress string
	connection       *net.UDPConn
	running          bool
	Consumer         func([]byte, *net.UDPAddr)
	mutex            sync.Mutex
	SkipInterfaces   []string
	Verbose          bool
}

func NewMulticastServer(multicastAddress string) (r *MulticastServer) {
	r = new(MulticastServer)
	r.multicastAddress = multicastAddress
	r.Consumer = func([]byte, *net.UDPAddr) {
		// noop by default
	}
	return
}

func (r *MulticastServer) Start() {
	r.running = true

	ifis := interfaces(r.SkipInterfaces)
	var ifiNames []string
	for _, ifi := range ifis {
		ifiNames = append(ifiNames, ifi.Name)
	}
	log.Printf("Listening on %s %s", r.multicastAddress, ifiNames)

	go r.receive()
}

func (r *MulticastServer) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = false
	if err := r.connection.Close(); err != nil {
		log.Println("Could not close connection: ", err)
	}
}

func (r *MulticastServer) receive() {
	var currentIfiIdx = 0
	for r.isRunning() {
		ifis := interfaces(r.SkipInterfaces)
		if len(ifis) > 0 {
			currentIfiIdx = currentIfiIdx % len(ifis)
			ifi := ifis[currentIfiIdx]
			r.receiveOnInterface(ifi)
			currentIfiIdx++
		} else {
			currentIfiIdx = 0
		}
		if currentIfiIdx >= len(ifis) {
			// cycled though all interfaces once, make a short break to avoid producing endless log messages
			time.Sleep(1 * time.Second)
		}
	}
}

func (r *MulticastServer) isRunning() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.running
}

func (r *MulticastServer) connectToInterface(ifi net.Interface) bool {
	addr, err := net.ResolveUDPAddr("udp", r.multicastAddress)
	if err != nil {
		log.Printf("Could resolve multicast address %v: %v", r.multicastAddress, err)
		return false
	}

	r.connection, err = net.ListenMulticastUDP("udp", &ifi, addr)
	if err != nil {
		log.Printf("Could not listen at %v on %v: %v", r.multicastAddress, ifi.Name, err)
		return false
	}

	if err := r.connection.SetReadBuffer(maxDatagramSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	if r.Verbose {
		log.Printf("Listening on %s (%s)", r.multicastAddress, ifi.Name)
	}

	return true
}

func (r *MulticastServer) receiveOnInterface(ifi net.Interface) {
	if !r.connectToInterface(ifi) {
		return
	}

	if r.Verbose {
		defer log.Printf("Stop listening on %s (%s)", r.multicastAddress, ifi.Name)
	}

	first := true
	data := make([]byte, maxDatagramSize)
	for {
		if err := r.connection.SetDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
			log.Println("Could not set deadline on connection: ", err)
		}
		n, remoteAddr, err := r.connection.ReadFromUDP(data)
		if err != nil {
			if r.Verbose {
				log.Println("ReadFromUDP failed:", err)
			}
			break
		}

		if first && r.Verbose {
			log.Printf("Got first data packets from %s (%s)", r.multicastAddress, ifi.Name)
			first = false
		}

		r.Consumer(data[:n], remoteAddr)
	}

	if r.Verbose {
		log.Printf("Stop listening on %s (%s)", r.multicastAddress, ifi.Name)
	}

	if err := r.connection.Close(); err != nil {
		log.Println("Could not close listener: ", err)
	}
	return
}
