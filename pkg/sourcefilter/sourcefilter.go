package sourcefilter

import (
	"log"
	"net"
	"sync"
	"time"
)

// Clock provides an interface for time operations to enable testing
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the actual system time
type RealClock struct{}

// Now returns the current system time
func (RealClock) Now() time.Time {
	return time.Now()
}

// SourceFilter filters messages based on source IP address with timeout-based failover
type SourceFilter struct {
	mu           sync.Mutex
	activeSource net.IP
	lastSeen     time.Time
	timeout      time.Duration
	clock        Clock
}

// New creates a new SourceFilter with the specified timeout.
func New(timeout time.Duration, clock Clock) *SourceFilter {
	return &SourceFilter{
		timeout: timeout,
		clock:   clock,
	}
}

// Accept checks if a message from the given source IP should be accepted.
// Returns true if the message should be processed, false if it should be rejected.
//
// Filtering logic:
// 1. First message from any source → lock onto that source
// 2. Same source → accept and refresh timestamp
// 3. Different source + active source timed out → switch to new source (log event)
// 4. Different source + active source alive → reject silently
func (sf *SourceFilter) Accept(sourceIP net.IP) bool {
	if sourceIP == nil {
		return true // Accept nil IPs (no filtering)
	}

	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := sf.clock.Now()

	// First message from any source - lock onto it
	if sf.activeSource == nil {
		sf.activeSource = sourceIP
		sf.lastSeen = now
		log.Printf("Source filter: locked onto source %s", sourceIP)
		return true
	}

	// Same source - accept and refresh timestamp
	if sf.activeSource.Equal(sourceIP) {
		sf.lastSeen = now
		return true
	}

	// Different source - check if active source has timed out
	timeSinceLastSeen := now.Sub(sf.lastSeen)
	if timeSinceLastSeen > sf.timeout {
		// Active source timed out - switch to new source
		log.Printf("Source filter: switching from %s to %s (timeout: %v)",
			sf.activeSource, sourceIP, timeSinceLastSeen)
		sf.activeSource = sourceIP
		sf.lastSeen = now
		return true
	}

	// Different source but active source still alive - reject silently
	return false
}

// ActiveSource returns the currently active source IP, or nil if none is active
func (sf *SourceFilter) ActiveSource() net.IP {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.activeSource
}

// Reset clears the active source, allowing the filter to accept a new source
func (sf *SourceFilter) Reset() {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.activeSource = nil
	sf.lastSeen = time.Time{}
}
