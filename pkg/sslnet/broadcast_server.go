package sslnet

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-reuseport"
)

// BroadcastServer receives UDP broadcast datagrams on a port. It mirrors the
// public API of MulticastServer so callers can swap between the two, but unlike
// multicast it needs no group membership: it binds a single socket to
// 0.0.0.0:<port> and receives all broadcast datagrams on that port.
type BroadcastServer struct {
	address    string
	connection *net.UDPConn
	running    bool
	Consumer   func([]byte, *net.UDPAddr)
	mutex      sync.Mutex
	// SkipInterfaces is retained for API parity with MulticastServer. It has no
	// effect on a broadcast server, which binds a single 0.0.0.0 socket.
	SkipInterfaces []string
	Verbose        bool
	NetworkServer
}

func NewBroadcastServer(address string) (r *BroadcastServer) {
	r = new(BroadcastServer)
	r.address = address
	r.Consumer = func([]byte, *net.UDPAddr) {
		// noop by default
	}
	return
}

func (r *BroadcastServer) Start() {
	r.running = true
	log.Printf("Listening for broadcast on %s", r.address)
	go r.receive()
}

func (r *BroadcastServer) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = false
	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			log.Println("Could not close connection: ", err)
		}
	}
}

func (r *BroadcastServer) receive() {
	for r.isRunning() {
		if !r.connect() {
			// Avoid a hot loop of failed connection attempts.
			time.Sleep(1 * time.Second)
			continue
		}
		r.receiveOnConnection()
	}
}

func (r *BroadcastServer) isRunning() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.running
}

func (r *BroadcastServer) connect() bool {
	addr, err := net.ResolveUDPAddr("udp", r.address)
	if err != nil {
		log.Printf("Could not resolve address %v: %v", r.address, err)
		return false
	}

	// Listen using go-reuseport which sets SO_REUSEADDR and SO_REUSEPORT
	listenAddr := &net.UDPAddr{IP: net.IPv4zero, Port: addr.Port}
	packetConn, err := reuseport.ListenPacket("udp4", listenAddr.String())
	if err != nil {
		log.Printf("Could not listen at %v: %v", r.address, err)
		return false
	}

	conn, ok := packetConn.(*net.UDPConn)
	if !ok {
		log.Printf("Could not cast to UDPConn")
		if err := packetConn.Close(); err != nil {
			log.Println("Could not close connection: ", err)
		}
		return false
	}

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	r.mutex.Lock()
	r.connection = conn
	r.mutex.Unlock()

	if r.Verbose {
		log.Printf("Listening for broadcast on %s", r.address)
	}

	return true
}

func (r *BroadcastServer) receiveOnConnection() {
	first := true
	data := make([]byte, maxDatagramSize)
	for {
		if err := r.connection.SetDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
			log.Println("Could not set deadline on connection: ", err)
		}
		n, remoteAddr, err := r.connection.ReadFromUDP(data)
		if err != nil {
			// A deadline timeout is expected while idle: re-arm and keep the
			// same socket open. Any other error (e.g. the closed connection
			// from Stop()) ends this receive loop.
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() && r.isRunning() {
				continue
			}
			if r.Verbose {
				log.Println("ReadFromUDP failed:", err)
			}
			break
		}

		if first && r.Verbose {
			log.Printf("Got first data packets from %s", r.address)
			first = false
		}

		r.Consumer(data[:n], remoteAddr)
	}

	if r.Verbose {
		log.Printf("Stop listening for broadcast on %s", r.address)
	}

	// If we are still running we broke out on a real read error and own the
	// cleanup before reconnecting. When stopped, Stop() already closed it.
	if r.isRunning() {
		if err := r.connection.Close(); err != nil {
			log.Println("Could not close listener: ", err)
		}
	}
}
