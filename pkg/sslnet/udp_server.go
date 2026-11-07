package sslnet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/go-reuseport"
	"golang.org/x/net/ipv4"
)

const (
	// receiveBufferSize is the requested socket receive buffer (SO_RCVBUF). Vision traffic from
	// several cameras at 60-75 Hz easily overflows a small buffer.
	receiveBufferSize = 2 * 1024 * 1024
	// readBufferSize is the size of the buffer a single datagram is read into.
	readBufferSize = 65536
	// interfaceCheckInterval is how often the eligible interfaces are re-checked, so that new
	// interfaces are joined and interfaces that went down are left (and re-joined once back up).
	interfaceCheckInterval = 5 * time.Second
	// reconnectDelay is the pause before a socket is recreated after a real socket error.
	reconnectDelay = 1 * time.Second
)

type receiveMode int

const (
	// modeMulticast joins the multicast group on every eligible interface.
	modeMulticast receiveMode = iota
	// modeBroadcast receives broadcast (and unicast) datagrams on the port without any
	// group membership.
	modeBroadcast
)

func (m receiveMode) String() string {
	if m == modeBroadcast {
		return "broadcast"
	}
	return "multicast"
}

// NetworkServer is implemented by UdpServer.
type NetworkServer interface {
	Start() error
	Stop() error
}

var _ NetworkServer = (*UdpServer)(nil)

// UdpServer receives UDP datagrams sent to an address. The kind of address decides how:
//
//   - A multicast address (e.g. "224.5.23.2:10006") is received via multicast: a single
//     long-lived socket joins the group on all eligible interfaces (up, multicast capable, with
//     an IPv4 address, not in SkipInterfaces) and stays joined. Read timeouts never cause a
//     leave/rejoin: they only drive a periodic interface re-check, which joins new interfaces and
//     re-joins interfaces that went down and up again. The socket is only recreated after a real
//     socket error.
//   - Any other address (e.g. "255.255.255.255:10006" or ":10006") is received via broadcast: no
//     group is joined, all datagrams to the port are received. This includes multicast datagrams
//     for the port, if the host is a member of the group (e.g. joined by another process).
//
// In both cases, the socket is bound to 0.0.0.0:<port> and receives from any interface. In
// multicast mode, datagrams not sent to the group (e.g. unicast or broadcast to the same port) are
// dropped, where the platform reports the destination address.
type UdpServer struct {
	address string
	mode    receiveMode

	// Consumer is called with each received datagram and its sender.
	Consumer func([]byte, *net.UDPAddr)
	// SkipInterfaces lists interface names to ignore. In multicast mode, the group is not
	// joined on these interfaces. In broadcast mode, datagrams received on them are dropped
	// (where the platform reports the receiving interface).
	SkipInterfaces []string
	// Verbose enables additional logging.
	Verbose bool

	mutex      sync.Mutex
	running    bool
	stop       chan struct{}
	connection *net.UDPConn
}

// NewUdpServer creates a server for the given address. A multicast address is received via
// multicast, any other address via broadcast. It fails if the address can not be resolved.
func NewUdpServer(address string) (*UdpServer, error) {
	if _, err := net.ResolveUDPAddr("udp4", address); err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", address, err)
	}
	return newUdpServer(address), nil
}

// newUdpServer creates a server without validating the address. An unresolvable address is
// reported (and retried) when the server is started.
func newUdpServer(address string) *UdpServer {
	mode := modeMulticast
	if addr, err := net.ResolveUDPAddr("udp4", address); err == nil && !addr.IP.IsMulticast() {
		mode = modeBroadcast
	}
	return &UdpServer{
		address: address,
		mode:    mode,
		Consumer: func([]byte, *net.UDPAddr) {
			// noop by default
		},
	}
}

// IsBroadcast reports whether the server receives via broadcast, because it was created with a
// non-multicast address.
func (r *UdpServer) IsBroadcast() bool {
	return r.mode == modeBroadcast
}

// Start starts receiving in the background. Starting a running server has no effect.
func (r *UdpServer) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.running {
		return nil
	}
	if r.mode == modeMulticast {
		log.Printf("Listening for multicast on %s %s", r.address, eligibleInterfaceNames(r.SkipInterfaces))
	} else {
		log.Printf("Listening for broadcast on %s", r.address)
	}
	r.running = true
	// Each run gets its own stop channel, so that a receive goroutine from a previous run can
	// never keep running after a quick Stop()/Start() sequence.
	r.stop = make(chan struct{})
	go r.receive(r.stop)
	return nil
}

// Stop stops receiving and closes the socket, which leaves all joined groups. Stopping a server
// that is not running has no effect.
func (r *UdpServer) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.running {
		return nil
	}
	r.running = false
	close(r.stop)
	if r.connection == nil {
		return nil
	}
	// Closing the socket unblocks the reader.
	err := r.connection.Close()
	r.connection = nil
	return err
}

func (r *UdpServer) receive(stop chan struct{}) {
	for {
		conn, err := r.listen()
		if err != nil {
			log.Printf("Could not listen for %s on %s: %v", r.mode, r.address, err)
		} else if r.setConnection(conn, stop) {
			r.serve(conn, stop)
			r.clearConnection(conn)
		} else {
			// Stopped while the socket was being created.
			_ = conn.Close()
			return
		}

		select {
		case <-stop:
			return
		case <-time.After(reconnectDelay):
		}
	}
}

// setConnection publishes conn as the current connection, unless the server was stopped.
func (r *UdpServer) setConnection(conn *net.UDPConn, stop chan struct{}) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	select {
	case <-stop:
		return false
	default:
	}
	r.connection = conn
	return true
}

// clearConnection closes conn after a socket error, unless Stop() already did.
func (r *UdpServer) clearConnection(conn *net.UDPConn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.connection != conn {
		return
	}
	if err := conn.Close(); err != nil {
		log.Println("Could not close connection: ", err)
	}
	r.connection = nil
}

func (r *UdpServer) currentConnection() *net.UDPConn {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.connection
}

func (r *UdpServer) listen() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", r.address)
	if err != nil {
		return nil, fmt.Errorf("could not resolve address: %w", err)
	}
	if r.mode == modeMulticast && !addr.IP.IsMulticast() {
		return nil, fmt.Errorf("%v is not a multicast address", addr.IP)
	}

	// Always bind to the wildcard address: Go rebinds a multicast address to the wildcard anyway,
	// and Windows does not allow binding to a multicast address at all. Datagrams that do not
	// belong to the mode are filtered by their destination address in serve().
	listenAddr := &net.UDPAddr{IP: net.IPv4zero, Port: addr.Port}

	// Listen using go-reuseport which sets SO_REUSEADDR and SO_REUSEPORT
	packetConn, err := reuseport.ListenPacket("udp4", listenAddr.String())
	if err != nil {
		return nil, err
	}
	conn, ok := packetConn.(*net.UDPConn)
	if !ok {
		_ = packetConn.Close()
		return nil, errors.New("could not cast to UDPConn")
	}

	if err := conn.SetReadBuffer(receiveBufferSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	if r.Verbose {
		log.Printf("Opened %s socket for %s at %s", r.mode, r.address, listenAddr)
	}
	return conn, nil
}

// serve reads from conn until a real socket error occurs or the server is stopped.
func (r *UdpServer) serve(conn *net.UDPConn, stop chan struct{}) {
	group, err := net.ResolveUDPAddr("udp4", r.address)
	if err != nil {
		log.Printf("Could not resolve address %v: %v", r.address, err)
		return
	}

	pc := ipv4.NewPacketConn(conn)
	if r.mode == modeMulticast {
		if err := pc.SetMulticastLoopback(true); err != nil {
			log.Printf("Could not set multicast loopback: %v", err)
		}
	}
	if err := pc.SetControlMessage(ipv4.FlagInterface|ipv4.FlagDst, true); err != nil && r.Verbose {
		// Not supported on all platforms (e.g. Windows). Datagrams are then not filtered by
		// destination and receiving interface.
		log.Printf("Could not enable interface control messages: %v", err)
	}

	m := membership{server: r, pc: pc, group: group, joined: map[int]string{}, names: map[int]string{}}
	if r.mode == modeMulticast {
		defer m.leaveAll()
	}
	m.refresh()
	nextCheck := time.Now().Add(interfaceCheckInterval)

	data := make([]byte, readBufferSize)
	firstFrom := map[string]bool{}
	for {
		if err := pc.SetReadDeadline(nextCheck); err != nil {
			log.Println("Could not set read deadline: ", err)
		}
		n, cm, src, err := pc.ReadFrom(data)

		if time.Now().After(nextCheck) {
			m.refresh()
			nextCheck = time.Now().Add(interfaceCheckInterval)
		}

		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				// No traffic within the check interval. Stay joined and keep the socket.
				continue
			}
			select {
			case <-stop:
				// Socket was closed by Stop()
			default:
				log.Printf("Could not read %s on %s, recreating socket: %v", r.mode, r.address, err)
			}
			return
		}

		ifiName := ""
		if cm != nil {
			if !r.accepts(cm.Dst, group.IP) {
				continue
			}
			ifiName = m.interfaceName(cm.IfIndex)
			if r.mode == modeBroadcast && r.skipped(ifiName) {
				continue
			}
		}

		if r.Verbose && !firstFrom[ifiName] {
			firstFrom[ifiName] = true
			if ifiName == "" {
				log.Printf("Got first %s data packets on %s", r.mode, r.address)
			} else {
				log.Printf("Got first %s data packets on %s (%s)", r.mode, r.address, ifiName)
			}
		}

		udpAddr, _ := src.(*net.UDPAddr)
		r.Consumer(data[:n], udpAddr)
	}
}

// accepts reports whether a datagram sent to dst belongs to this server. In multicast mode, only
// datagrams sent to the group are accepted. In broadcast mode, everything is accepted.
func (r *UdpServer) accepts(dst net.IP, group net.IP) bool {
	if dst == nil || r.mode != modeMulticast {
		return true
	}
	return dst.Equal(group)
}

func (r *UdpServer) skipped(ifiName string) bool {
	return slices.Contains(r.SkipInterfaces, ifiName)
}

// membership keeps track of the interfaces the group is joined on.
type membership struct {
	server *UdpServer
	pc     *ipv4.PacketConn
	group  *net.UDPAddr
	// joined maps the index of each joined interface to its name
	joined map[int]string
	// names maps interface indexes to names, for logging
	names map[int]string
	// warnedNoInterface avoids repeating the "no interface" log message on every check
	warnedNoInterface bool
}

// refresh updates the interface name cache and, in multicast mode, joins the group on new
// eligible interfaces and leaves it on interfaces that are gone or no longer eligible (e.g. down).
// An interface that goes down and up again is thereby re-joined.
func (m *membership) refresh() {
	if all, err := net.Interfaces(); err == nil {
		m.names = map[int]string{}
		for _, ifi := range all {
			m.names[ifi.Index] = ifi.Name
		}
	}

	if m.server.mode != modeMulticast {
		return
	}

	eligible := map[int]net.Interface{}
	for _, ifi := range interfaces(m.server.SkipInterfaces) {
		eligible[ifi.Index] = ifi
	}

	for index, name := range m.joined {
		if ifi, ok := eligible[index]; ok && ifi.Name == name {
			continue
		}
		// Interface is gone, down or now skipped. Leaving may fail if the interface no longer
		// exists, which is fine.
		ifi := net.Interface{Index: index, Name: name}
		_ = m.pc.LeaveGroup(&ifi, m.group)
		delete(m.joined, index)
		log.Printf("Left multicast group %s on %s", m.server.address, name)
	}

	for index, ifi := range eligible {
		if _, ok := m.joined[index]; ok {
			continue
		}
		err := m.pc.JoinGroup(&ifi, m.group)
		if err != nil && !errors.Is(err, syscall.EADDRINUSE) {
			// EADDRINUSE: still joined from before (e.g. a leave failed), so treat as joined.
			log.Printf("Could not join multicast group %s on %s: %v", m.server.address, ifi.Name, err)
			continue
		}
		m.joined[index] = ifi.Name
		log.Printf("Joined multicast group %s on %s", m.server.address, ifi.Name)
	}

	if len(m.joined) == 0 {
		if !m.warnedNoInterface {
			log.Printf("Multicast group %s not joined on any interface yet, retrying every %v",
				m.server.address, interfaceCheckInterval)
			m.warnedNoInterface = true
		}
	} else {
		m.warnedNoInterface = false
	}
}

func (m *membership) leaveAll() {
	for index, name := range m.joined {
		ifi := net.Interface{Index: index, Name: name}
		_ = m.pc.LeaveGroup(&ifi, m.group)
	}
	m.joined = map[int]string{}
}

func (m *membership) interfaceName(index int) string {
	if name, ok := m.names[index]; ok {
		return name
	}
	if ifi, err := net.InterfaceByIndex(index); err == nil {
		m.names[index] = ifi.Name
		return ifi.Name
	}
	return ""
}

func eligibleInterfaceNames(skipInterfaces []string) (names []string) {
	for _, ifi := range interfaces(skipInterfaces) {
		names = append(names, ifi.Name)
	}
	return
}
