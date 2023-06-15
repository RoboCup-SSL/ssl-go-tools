package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const maxDatagramSize = 8192

type Source struct {
	nif  string
	ip   string
	port int
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

	ifiList := interfaces()
	for _, address := range mcAddresses {
		for _, ifi := range ifiList {
			go receiveOnInterface(address, ifi)
		}
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

func receiveOnInterface(address string, ifi net.Interface) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Printf("Could resolve multicast address %v: %v", address, err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp", &ifi, addr)
	if err != nil {
		log.Printf("Could not listen at %v: %v", address, err)
		return
	}

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	log.Printf("Listening on %s (%s)", address, ifi.Name)

	data := make([]byte, maxDatagramSize)
	for {
		_, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Println("ReadFromUDP failed:", err)
			return
		}

		addRemote(address, Source{
			nif:  ifi.Name,
			ip:   remoteAddr.IP.String(),
			port: remoteAddr.Port,
		})
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
	log.Printf("New source on %v: %+v\n", address, source)
}
