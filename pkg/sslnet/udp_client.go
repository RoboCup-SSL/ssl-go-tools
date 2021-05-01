package sslnet

import (
	"log"
	"net"
	"sync"
)

// UdpClient establishes a UDP connection to a server
type UdpClient struct {
	Name      string
	Consumer  func([]byte)
	address   string
	conns     []*net.UDPConn
	running   bool
	mutex     sync.Mutex
	receivers sync.WaitGroup
}

// NewUdpClient creates a new UDP client
func NewUdpClient(address string) (t *UdpClient) {
	t = new(UdpClient)
	t.Name = "UdpClient"
	t.address = address
	t.Consumer = func([]byte) {}
	return
}

// Start the client by listening for responses it a separate goroutine
func (c *UdpClient) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.running {
		log.Printf("%v - Starting", c.Name)
		c.running = true
		c.connect()
		log.Printf("%v - Started", c.Name)
	}
}

// Stop the client by stop listening for responses and closing all existing connections
func (c *UdpClient) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.running {
		log.Printf("%v - Stopping", c.Name)
		c.running = false
		for _, conn := range c.conns {
			if err := conn.Close(); err != nil {
				log.Printf("%v - Could not close client connection: %v", c.Name, err)
			}
		}
		c.receivers.Wait()
		c.conns = []*net.UDPConn{}
		log.Printf("%v - Stopped", c.Name)
	}
}

// Send data to the server
func (c *UdpClient) Send(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, conn := range c.conns {
		if _, err := conn.Write(data); err != nil {
			log.Printf("%v - Could not write to %s at %s: %s", c.Name, conn.RemoteAddr(), conn.LocalAddr(), err)
		}
	}
}

func (c *UdpClient) connect() {
	log.Printf("%v - Connecting to %v", c.Name, c.address)
	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		log.Printf("%v - Could resolve address %v: %v", c.Name, c.address, err)
		return
	}

	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("%v - Could not retrieve interface addresses: %v", c.Name, err)
		return
	}

	for _, iaddr := range iaddrs {
		ip := iaddr.(*net.IPNet).IP
		if ip.To4() == nil {
			// Ignore IPv6 for now
			continue
		}
		laddr := &net.UDPAddr{IP: ip}
		conn, err := net.DialUDP("udp", laddr, addr)
		if err != nil {
			log.Printf("%v - Could not connect to %v at %v: %v", c.Name, addr, laddr, err)
			continue
		}

		if err := conn.SetWriteBuffer(maxDatagramSize); err != nil {
			log.Printf("%v - Could not set read buffer: %v", c.Name, err)
		}

		c.conns = append(c.conns, conn)
		go c.receive(conn)
	}
}

func (c *UdpClient) receive(conn *net.UDPConn) {
	log.Printf("%v - Connected to %s at %s", c.Name, conn.RemoteAddr(), conn.LocalAddr())
	defer log.Printf("%v - Disconnected from %s at %s", c.Name, conn.RemoteAddr(), conn.LocalAddr())

	c.receivers.Add(1)
	defer c.receivers.Done()

	data := make([]byte, maxDatagramSize)
	for {
		n, _, err := conn.ReadFrom(data)
		if err != nil {
			if opErr, ok := err.(*net.OpError); !ok || opErr.Err.Error() != "use of closed network connection" {
				log.Printf("%v - Could not receive data from %s at %s: %s", c.Name, conn.RemoteAddr(), conn.LocalAddr(), err)
			}
			return
		} else {
			c.Consumer(data[:n])
		}
	}
}
