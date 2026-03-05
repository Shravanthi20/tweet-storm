package algorithms

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"sync"
)

// HashRing manages a consistent hashing ring for node distribution
type HashRing struct {
	mu           sync.RWMutex
	replicas     int
	circle       map[uint32]string
	sortedHashes []uint32
}

// NewHashRing creates a new HashRing with the specified number of virtual nodes (replicas)
func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		replicas: replicas,
		circle:   make(map[uint32]string),
	}
}

// hashKey generates a uint32 hash for a given string key
// using SHA-256 for good distribution
func (hr *HashRing) hashKey(key string) uint32 {
	h := sha256.New()
	h.Write([]byte(key))
	sum := h.Sum(nil)

	// Take the first 4 bytes to create a uint32
	return binary.BigEndian.Uint32(sum[:4])
}

// AddNode adds a new node (like a worker URL) to the hash ring
func (hr *HashRing) AddNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Check if already in the ring (simplified check)
	for _, existingNode := range hr.circle {
		if existingNode == node {
			return // already present
		}
	}

	// Add virtual nodes (replicas) to the ring
	for i := 0; i < hr.replicas; i++ {
		// Create a unique key for each replica of the node
		replicaKey := string(rune(i)) + node
		hash := hr.hashKey(replicaKey)

		hr.circle[hash] = node
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}

	// Keep the ring sorted for binary search
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

// RemoveNode safely removes a node and all its replicas from the ring
func (hr *HashRing) RemoveNode(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Find the hashes to remove
	var hashesToRemove []uint32
	for hash, ringNode := range hr.circle {
		if ringNode == node {
			hashesToRemove = append(hashesToRemove, hash)
		}
	}

	// Remove from the map
	for _, hash := range hashesToRemove {
		delete(hr.circle, hash)
	}

	// Rebuild the sortedHashes slice without the removed hashes
	hr.updateSortedHashes()
}

// updateSortedHashes rebuilds the sortedHashes slice from the current circle map
// Note: caller must hold the write lock
func (hr *HashRing) updateSortedHashes() {
	hr.sortedHashes = make([]uint32, 0, len(hr.circle))
	for hash := range hr.circle {
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

// GetNode returns the closest node in the ring for the provided key
func (hr *HashRing) GetNode(key string) string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.circle) == 0 {
		return ""
	}

	hash := hr.hashKey(key)

	// Binary search to find the nearest higher hash on the ring
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	// Wrap around if the hash is greater than the largest hash in the ring
	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.circle[hr.sortedHashes[idx]]
}

// GetNodes returns a list of all unique nodes currently in the ring
func (hr *HashRing) GetNodes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	uniqueNodes := make(map[string]bool)
	var nodes []string

	for _, node := range hr.circle {
		if !uniqueNodes[node] {
			uniqueNodes[node] = true
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// Size returns the number of distinct physical nodes in the ring
func (hr *HashRing) Size() int {
	return len(hr.GetNodes())
}
