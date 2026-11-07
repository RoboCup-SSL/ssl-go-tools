package sslnet

import (
	"log"
)

const maxDatagramSize = 8192

// MulticastServer receives UDP multicast datagrams.
//
// It is kept for compatibility and delegates to UdpServer. Like UdpServer, it receives via
// broadcast if created with a non-multicast address.
//
// Deprecated: Use UdpServer instead.
type MulticastServer struct {
	*UdpServer
	NetworkServer
}

// Deprecated: Use NewUdpServer instead.
func NewMulticastServer(multicastAddress string) (r *MulticastServer) {
	return &MulticastServer{UdpServer: newUdpServer(multicastAddress)}
}

func (r *MulticastServer) Start() {
	_ = r.UdpServer.Start()
}

func (r *MulticastServer) Stop() {
	if err := r.UdpServer.Stop(); err != nil {
		log.Println("Could not close connection: ", err)
	}
}
