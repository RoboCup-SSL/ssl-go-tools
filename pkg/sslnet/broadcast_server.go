package sslnet

import (
	"log"
	"net"
	"strconv"
)

// BroadcastServer receives UDP broadcast datagrams on a port. The host part of the address is
// ignored.
//
// It is kept for compatibility and delegates to UdpServer.
//
// Deprecated: Use UdpServer with a broadcast address (e.g. "255.255.255.255:10006" or ":10006")
// instead.
type BroadcastServer struct {
	*UdpServer
	NetworkServer
}

// Deprecated: Use NewUdpServer with a broadcast address instead.
func NewBroadcastServer(address string) (r *BroadcastServer) {
	// Drop the host, so that UdpServer receives via broadcast even if it is a multicast address.
	if _, port, err := net.SplitHostPort(address); err == nil {
		if _, err := strconv.Atoi(port); err == nil {
			address = ":" + port
		}
	}
	return &BroadcastServer{UdpServer: newUdpServer(address)}
}

func (r *BroadcastServer) Start() {
	_ = r.UdpServer.Start()
}

func (r *BroadcastServer) Stop() {
	if err := r.UdpServer.Stop(); err != nil {
		log.Println("Could not close connection: ", err)
	}
}
