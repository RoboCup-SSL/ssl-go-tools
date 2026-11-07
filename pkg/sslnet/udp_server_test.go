package sslnet

import (
	"net"
	"runtime"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/ipv4"
)

const testGroup = "224.5.23.250"

// testInterface returns an interface that is eligible for multicast, or skips the test.
func testInterface(t *testing.T) net.Interface {
	t.Helper()
	ifis := interfaces(nil)
	if len(ifis) == 0 {
		t.Skip("no multicast capable interface available")
	}
	return ifis[0]
}

func freePort(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	return conn.LocalAddr().(*net.UDPAddr).Port
}

// sendMulticast sends a multicast datagram out of ifi, looped back to the local host.
func sendMulticast(t *testing.T, ifi net.Interface, address string, payload string) {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	pc := ipv4.NewPacketConn(conn)
	if err := pc.SetMulticastInterface(&ifi); err != nil {
		t.Fatal(err)
	}
	if err := pc.SetMulticastLoopback(true); err != nil {
		t.Fatal(err)
	}
	dst, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pc.WriteTo([]byte(payload), nil, dst); err != nil {
		t.Fatal(err)
	}
}

func sendUnicast(t *testing.T, address string, payload string) {
	t.Helper()
	conn, err := net.Dial("udp4", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
}

func newTestServer(t *testing.T, address string) *UdpServer {
	t.Helper()
	server, err := NewUdpServer(address)
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func receiveChannel(s *UdpServer) chan string {
	received := make(chan string, 100)
	s.Consumer = func(data []byte, _ *net.UDPAddr) {
		received <- string(data)
	}
	return received
}

// waitForConnection waits until the server opened its socket and joined the group.
func waitForConnection(t *testing.T, s *UdpServer) *net.UDPConn {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if conn := s.currentConnection(); conn != nil {
			// give the receive goroutine time to join the group
			time.Sleep(100 * time.Millisecond)
			return conn
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not open a connection")
	return nil
}

// sendUntilReceived resends until the payload arrives, to tolerate a slow first join.
func sendUntilReceived(t *testing.T, send func(), received chan string, payload string) {
	t.Helper()
	for range 20 {
		send()
		select {
		case got := <-received:
			if got != payload {
				t.Fatalf("received %q, want %q", got, payload)
			}
			// drain duplicates from earlier attempts
			for {
				select {
				case <-received:
				case <-time.After(200 * time.Millisecond):
					return
				}
			}
		case <-time.After(100 * time.Millisecond):
		}
	}
	t.Fatalf("did not receive %q", payload)
}

func expectNothing(t *testing.T, received chan string, d time.Duration) {
	t.Helper()
	select {
	case got := <-received:
		t.Fatalf("unexpectedly received %q", got)
	case <-time.After(d):
	}
}

func TestUdpServer_ReceivesAndStaysJoinedDuringGaps(t *testing.T) {
	ifi := testInterface(t)
	address := testGroup + ":" + strconv.Itoa(freePort(t))

	server := newTestServer(t, address)
	if server.IsBroadcast() {
		t.Fatal("expected multicast mode")
	}
	received := receiveChannel(server)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Stop() }()
	conn := waitForConnection(t, server)

	sendUntilReceived(t, func() { sendMulticast(t, ifi, address, "first") }, received, "first")

	// A gap much longer than the old 300 ms timeout must not close the socket (and thereby
	// leave the group).
	time.Sleep(1500 * time.Millisecond)
	if server.currentConnection() != conn {
		t.Fatal("socket was recreated during a gap in traffic")
	}

	sendMulticast(t, ifi, address, "second")
	select {
	case got := <-received:
		if got != "second" {
			t.Fatalf("received %q, want %q", got, "second")
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive datagram after gap")
	}
}

func TestUdpServer_StopAndRestart(t *testing.T) {
	ifi := testInterface(t)
	address := testGroup + ":" + strconv.Itoa(freePort(t))

	server := newTestServer(t, address)
	// Stop before Start must not panic
	_ = server.Stop()

	received := receiveChannel(server)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	waitForConnection(t, server)
	_ = server.Stop()
	if server.currentConnection() != nil {
		t.Fatal("connection not cleared on stop")
	}

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Stop() }()
	waitForConnection(t, server)
	sendUntilReceived(t, func() { sendMulticast(t, ifi, address, "again") }, received, "again")
}

func TestUdpServer_BroadcastAddress(t *testing.T) {
	port := freePort(t)
	server := newTestServer(t, "255.255.255.255:"+strconv.Itoa(port))
	if !server.IsBroadcast() {
		t.Fatal("expected broadcast mode for a non-multicast address")
	}
	received := receiveChannel(server)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Stop() }()
	waitForConnection(t, server)

	target := "127.0.0.1:" + strconv.Itoa(port)
	sendUntilReceived(t, func() { sendUnicast(t, target, "broadcast") }, received, "broadcast")
}

func TestUdpServer_MulticastAndBroadcastOnSamePort(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("destination filtering relies on control messages, verified on Linux only")
	}
	ifi := testInterface(t)
	port := strconv.Itoa(freePort(t))
	mcAddress := testGroup + ":" + port

	mcServer := newTestServer(t, mcAddress)
	mcReceived := receiveChannel(mcServer)
	_ = mcServer.Start()
	defer func() { _ = mcServer.Stop() }()
	waitForConnection(t, mcServer)

	bcServer := newTestServer(t, "255.255.255.255:"+port)
	if !bcServer.IsBroadcast() {
		t.Fatal("expected broadcast mode")
	}
	bcReceived := receiveChannel(bcServer)
	_ = bcServer.Start()
	defer func() { _ = bcServer.Stop() }()
	waitForConnection(t, bcServer)

	// Multicast reaches the multicast server, and the broadcast server, too, as the host is a
	// member of the group
	sendUntilReceived(t, func() { sendMulticast(t, ifi, mcAddress, "mc") }, mcReceived, "mc")
	sendUntilReceived(t, func() {}, bcReceived, "mc")

	// Non-multicast traffic reaches only the broadcast server
	sendUntilReceived(t, func() { sendUnicast(t, "127.0.0.1:"+port, "bc") }, bcReceived, "bc")
	expectNothing(t, mcReceived, 300*time.Millisecond)
}

func TestUdpServer_InvalidAddress(t *testing.T) {
	if _, err := NewUdpServer("not-an-address"); err == nil {
		t.Fatal("expected an error for an invalid address")
	}
}

// The compatibility wrappers keep their old API and delegate to UdpServer.
func TestCompatibilityWrappers(t *testing.T) {
	port := strconv.Itoa(freePort(t))

	mc := NewMulticastServer(testGroup + ":" + port)
	if mc.IsBroadcast() {
		t.Fatal("expected multicast mode")
	}

	// BroadcastServer ignores the host, even if it is a multicast address
	bc := NewBroadcastServer(testGroup + ":" + port)
	if !bc.IsBroadcast() {
		t.Fatal("expected broadcast mode")
	}
	received := make(chan string, 10)
	bc.Consumer = func(data []byte, _ *net.UDPAddr) { received <- string(data) }
	bc.SkipInterfaces = []string{}
	bc.Verbose = true
	bc.Start()
	defer bc.Stop()
	waitForConnection(t, bc.UdpServer)
	sendUntilReceived(t, func() { sendUnicast(t, "127.0.0.1:"+port, "compat") }, received, "compat")

	// Stop before Start must not panic
	mc.Stop()
}
