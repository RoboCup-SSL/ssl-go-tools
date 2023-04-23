package main

import (
	"flag"
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

var detectedRemotes map[string][]string

func main() {
	flag.Parse()

	detectedRemotes = map[string][]string{}
	mcAddresses := flag.Args()
	if len(mcAddresses) == 0 {
		mcAddresses = []string{"224.5.23.1:10003", "224.5.23.2:10006", "224.5.23.2:10010", "224.5.23.2:10012"}
	}

	for _, address := range mcAddresses {
		go watchAddress(address)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

func watchAddress(address string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("Could not set read buffer to %v.", maxDatagramSize)
	}
	log.Println("Receiving from", address)
	for {
		_, udpAddr, err := conn.ReadFromUDP([]byte{0})
		if err != nil {
			log.Print("Could not read: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		addRemote(address, udpAddr.IP.String())
	}
}

func addRemote(address string, remote string) {
	remotes, ok := detectedRemotes[address]
	if !ok {
		detectedRemotes[address] = []string{}
	}
	for _, a := range remotes {
		if a == remote {
			return
		}
	}

	detectedRemotes[address] = append(detectedRemotes[address], remote)
	log.Printf("remote ip on %v: %v\n", address, remote)
}
