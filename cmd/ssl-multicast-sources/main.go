package main

import (
	"flag"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const maxDatagramSize = 8192

type Source struct {
	cm  string
	src string
}

var detectedRemotes map[string][]Source
var detectedRemotesMutex sync.Mutex

func main() {
	flag.Parse()

	detectedRemotes = map[string][]Source{}
	mcAddresses := flag.Args()
	if len(mcAddresses) == 0 {
		mcAddresses = []string{"224.5.23.1:10003", "224.5.23.2:10006", "224.5.23.2:10010", "224.5.23.2:10012"}
	}

	for _, address := range mcAddresses {
		go receiveOnInterfaceIpv4(address)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

func interfaces() (interfaces []net.Interface) {
	interfaces = []net.Interface{}
	ifis, err := net.Interfaces()
	if err != nil {
		log.Println("Could not get available interfaces: ", err)
	}
	for _, ifi := range ifis {
		interfaces = append(interfaces, ifi)
	}
	return
}

func receiveOnInterfaceIpv4(address string) {
	group := net.ParseIP(strings.Split(address, ":")[0])
	c, err := net.ListenPacket("udp4", address)
	if err != nil {
		log.Printf("Could not start listening on %v: %v", address, err)
		return
	}
	defer func(c net.PacketConn) {
		if err := c.Close(); err != nil {
			log.Println("Failed to close connection:", err)
		}
	}(c)

	log.Printf("Listening on %s", address)

	p := ipv4.NewPacketConn(c)

	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Println("Failed to set control message flag 'FlagDst':", err)
	}

	ifiList := interfaces()
	for _, ifi := range ifiList {
		if err := p.JoinGroup(&ifi, &net.UDPAddr{IP: group}); err != nil {
			log.Printf("Failed to join multicast group %v: %v", group, err)
		} else {
			log.Printf("Joined multicast group %v on interface %+v", group, ifi)
		}
	}

	data := make([]byte, maxDatagramSize)
	for {
		_, cm, src, err := p.ReadFrom(data)
		if err != nil {
			log.Printf("Could not read from %v: %v", address, err)
			continue
		}
		if cm.Dst.IsMulticast() {
			if cm.Dst.Equal(group) {
				addRemote(address, Source{
					cm:  cm.String(),
					src: src.String(),
				})
			} else {
				// unknown group, discard
				continue
			}
		}
	}
}

func addRemote(address string, source Source) {
	detectedRemotesMutex.Lock()
	defer detectedRemotesMutex.Unlock()

	remotes, ok := detectedRemotes[address]
	if !ok {
		detectedRemotes[address] = []Source{}
	}
	for _, a := range remotes {
		if a == source {
			return
		}
	}

	detectedRemotes[address] = append(detectedRemotes[address], source)
	log.Printf("New source on %v: %+v", address, source)
}
