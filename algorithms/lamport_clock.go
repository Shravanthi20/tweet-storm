package algorithms

import "sync"

// LamportClock tracks logical time for distributed event ordering.
type LamportClock struct {
	mu     sync.Mutex
	clock  int
	nodeID string
}

func NewLamportClock(nodeID string) *LamportClock {
	return &LamportClock{
		nodeID: nodeID,
	}
}

// Tick is used for local/internal events.
func (lc *LamportClock) Tick() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.clock++
	return lc.clock
}

// SendEvent increments logical time before a send.
func (lc *LamportClock) SendEvent() int {
	return lc.Tick()
}

// ReceiveEvent applies Lamport receive rule:
// clock = max(local_clock, received_timestamp) + 1
func (lc *LamportClock) ReceiveEvent(receivedTimestamp int) int {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if receivedTimestamp > lc.clock {
		lc.clock = receivedTimestamp
	}
	lc.clock++
	return lc.clock
}

func (lc *LamportClock) GetTime() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	return lc.clock
}
