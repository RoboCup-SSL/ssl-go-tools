package sourcefilter

import (
	"net"
	"sync"
	"testing"
	"time"
)

// MockClock is a controllable clock for testing
type MockClock struct {
	mu   sync.Mutex
	time time.Time
}

// Now returns the current mock time
func (m *MockClock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.time
}

// Advance moves the mock clock forward by the specified duration
func (m *MockClock) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.time = m.time.Add(d)
}

// SetTime sets the mock clock to a specific time
func (m *MockClock) SetTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.time = t
}

func TestFirstMessageAcceptance(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")

	// First message should be accepted
	if !filter.Accept(ip1) {
		t.Error("First message should be accepted")
	}

	// Active source should be set
	if !filter.ActiveSource().Equal(ip1) {
		t.Errorf("Active source should be %s, got %s", ip1, filter.ActiveSource())
	}
}

func TestSameSourceAcceptance(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")

	// Lock onto first source
	filter.Accept(ip1)

	// Advance time slightly
	clock.Advance(100 * time.Millisecond)

	// Same source should be accepted
	if !filter.Accept(ip1) {
		t.Error("Same source should be accepted")
	}

	// Advance more time
	clock.Advance(200 * time.Millisecond)

	// Same source should still be accepted
	if !filter.Accept(ip1) {
		t.Error("Same source should still be accepted")
	}
}

func TestDifferentSourceRejection(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	// Lock onto first source
	filter.Accept(ip1)

	// Advance time but not enough to timeout
	clock.Advance(300 * time.Millisecond)

	// Different source should be rejected
	if filter.Accept(ip2) {
		t.Error("Different source should be rejected when active source is alive")
	}

	// Active source should still be ip1
	if !filter.ActiveSource().Equal(ip1) {
		t.Errorf("Active source should still be %s, got %s", ip1, filter.ActiveSource())
	}
}

func TestTimeoutBasedSwitching(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	// Lock onto first source
	filter.Accept(ip1)

	// Advance time past timeout
	clock.Advance(600 * time.Millisecond)

	// Different source should be accepted after timeout
	if !filter.Accept(ip2) {
		t.Error("Different source should be accepted after timeout")
	}

	// Active source should now be ip2
	if !filter.ActiveSource().Equal(ip2) {
		t.Errorf("Active source should be %s, got %s", ip2, filter.ActiveSource())
	}
}

func TestTimeoutRefresh(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	// Lock onto first source
	filter.Accept(ip1)

	// Advance time
	clock.Advance(300 * time.Millisecond)

	// Accept from same source - should refresh timestamp
	filter.Accept(ip1)

	// Advance another 300ms (total 600ms from start, but only 300ms from refresh)
	clock.Advance(300 * time.Millisecond)

	// Different source should still be rejected (refresh extended the timeout)
	if filter.Accept(ip2) {
		t.Error("Different source should be rejected because timeout was refreshed")
	}

	// Advance past new timeout
	clock.Advance(300 * time.Millisecond)

	// Now different source should be accepted
	if !filter.Accept(ip2) {
		t.Error("Different source should be accepted after refreshed timeout")
	}
}

func TestConcurrentAccess(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	var wg sync.WaitGroup
	iterations := 1000

	// Multiple goroutines calling Accept concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if id%2 == 0 {
					filter.Accept(ip1)
				} else {
					filter.Accept(ip2)
				}
			}
		}(i)
	}

	// Multiple goroutines reading ActiveSource concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				filter.ActiveSource()
			}
		}()
	}

	wg.Wait()

	// Should not panic and should have an active source
	if filter.ActiveSource() == nil {
		t.Error("Should have an active source after concurrent access")
	}
}

func TestResetFunctionality(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	// Lock onto first source
	filter.Accept(ip1)

	if filter.ActiveSource() == nil {
		t.Error("Should have an active source")
	}

	// Reset the filter
	filter.Reset()

	// Active source should be nil
	if filter.ActiveSource() != nil {
		t.Error("Active source should be nil after reset")
	}

	// Should accept the next source immediately
	if !filter.Accept(ip2) {
		t.Error("Should accept new source after reset")
	}

	if !filter.ActiveSource().Equal(ip2) {
		t.Errorf("Active source should be %s after reset, got %s", ip2, filter.ActiveSource())
	}
}

func TestActiveSourceObservability(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	// Initially no active source
	if filter.ActiveSource() != nil {
		t.Error("Should have no active source initially")
	}

	ip1 := net.ParseIP("192.168.1.1")
	filter.Accept(ip1)

	// Should show active source
	if !filter.ActiveSource().Equal(ip1) {
		t.Error("Should show correct active source")
	}

	// Advance past timeout
	clock.Advance(600 * time.Millisecond)

	ip2 := net.ParseIP("192.168.1.2")
	filter.Accept(ip2)

	// Should show new active source
	if !filter.ActiveSource().Equal(ip2) {
		t.Error("Should show new active source after switch")
	}
}

func TestNilIPAcceptance(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	// Nil IP should be accepted (no filtering)
	if !filter.Accept(nil) {
		t.Error("Nil IP should be accepted")
	}

	// Active source should still be nil
	if filter.ActiveSource() != nil {
		t.Error("Active source should be nil when only nil IPs are accepted")
	}

	// Real IP after nil should be accepted
	ip1 := net.ParseIP("192.168.1.1")
	if !filter.Accept(ip1) {
		t.Error("Real IP should be accepted after nil")
	}
}

func TestIPv4VsIPv6(t *testing.T) {
	clock := &MockClock{time: time.Now()}
	filter := New(500*time.Millisecond, clock)

	ipv4 := net.ParseIP("192.168.1.1")
	ipv6 := net.ParseIP("2001:db8::1")

	// Lock onto IPv4
	filter.Accept(ipv4)

	// IPv6 should be rejected (different source)
	if filter.Accept(ipv6) {
		t.Error("IPv6 should be rejected when IPv4 is active")
	}

	// Advance past timeout
	clock.Advance(600 * time.Millisecond)

	// IPv6 should be accepted after timeout
	if !filter.Accept(ipv6) {
		t.Error("IPv6 should be accepted after timeout")
	}

	// Active source should be IPv6
	if !filter.ActiveSource().Equal(ipv6) {
		t.Error("Active source should be IPv6")
	}
}
