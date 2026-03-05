package algorithms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type RAState int

const (
	IDLE RAState = iota
	WAITING
	CRITICAL_SECTION
)

type RAMessage struct {
	NodeID    int `json:"nodeId"`
	Timestamp int `json:"timestamp"`
}

type RicartAgrawala struct {
	nodeID          int
	nodePort        string
	state           RAState
	clock           *LamportClock
	requestTime     int
	replyCount      int
	expectedReplies int
	deferredReplies []int
	mu              sync.Mutex
	cond            *sync.Cond
	nodes           map[int]string
}

func NewRicartAgrawala(nodeID int, port string, clock *LamportClock, workerNodes map[int]string) *RicartAgrawala {
	ra := &RicartAgrawala{
		nodeID:          nodeID,
		nodePort:        port,
		state:           IDLE,
		clock:           clock,
		expectedReplies: len(workerNodes),
		deferredReplies: make([]int, 0),
		nodes:           workerNodes,
	}
	ra.cond = sync.NewCond(&ra.mu)

	// Since we don't have a specific router, we just register globally.
	// Note: We'll modify the worker to actually serve these routes.
	http.HandleFunc("/ra/request", ra.handleRequest)
	http.HandleFunc("/ra/reply", ra.handleReply)
	http.HandleFunc("/ra/status", ra.handleStatus)

	return ra
}

func (ra *RicartAgrawala) RequestLock() {
	ra.mu.Lock()
	ra.state = WAITING
	ra.requestTime = ra.clock.SendEvent()
	ra.replyCount = 0
	
	fmt.Printf("[RA Node %d] Requesting Lock at logical time %d\n", ra.nodeID, ra.requestTime)
	
	if ra.expectedReplies == 0 {
		ra.state = CRITICAL_SECTION
		ra.mu.Unlock()
		return
	}
	
	reqTime := ra.requestTime
	myID := ra.nodeID
	targets := make(map[int]string)
	for id, url := range ra.nodes {
		targets[id] = url
	}
	ra.mu.Unlock()

	payload := RAMessage{NodeID: myID, Timestamp: reqTime}
	data, _ := json.Marshal(payload)

	for _, url := range targets {
		go func(targetURL string) {
			client := http.Client{Timeout: 2 * time.Second}
			_, err := client.Post(targetURL+"/ra/request", "application/json", bytes.NewBuffer(data))
			if err != nil {
				fmt.Printf("[RA Node %d] Failed to reach %s for lock. Implicit reply.\n", myID, targetURL)
				ra.addReply()
			}
		}(url)
	}

	ra.mu.Lock()
	for ra.replyCount < ra.expectedReplies {
		ra.cond.Wait()
	}
	ra.state = CRITICAL_SECTION
	fmt.Printf("[RA Node %d] ENTERING CRITICAL SECTION\n", ra.nodeID)
	ra.mu.Unlock()
}

func (ra *RicartAgrawala) ReleaseLock() {
	ra.mu.Lock()
	ra.state = IDLE
	fmt.Printf("[RA Node %d] EXITING CRITICAL SECTION. Sending %d deferred replies.\n", ra.nodeID, len(ra.deferredReplies))
	
	deferredToProcess := make([]int, len(ra.deferredReplies))
	copy(deferredToProcess, ra.deferredReplies)
	ra.deferredReplies = make([]int, 0)
	
	myID := ra.nodeID
	targets := ra.nodes
	ra.mu.Unlock()

	for _, targetID := range deferredToProcess {
		url, exists := targets[targetID]
		if exists {
			go func(tURL string) {
				payload := RAMessage{NodeID: myID, Timestamp: 0}
				data, _ := json.Marshal(payload)
				client := http.Client{Timeout: 1 * time.Second}
				client.Post(tURL+"/ra/reply", "application/json", bytes.NewBuffer(data))
			}(url)
		}
	}
}

func (ra *RicartAgrawala) handleRequest(w http.ResponseWriter, r *http.Request) {
	var req RAMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ra.clock.ReceiveEvent(req.Timestamp)

	ra.mu.Lock()
	defer ra.mu.Unlock()

	deferReply := false
	if ra.state == CRITICAL_SECTION {
		deferReply = true
	} else if ra.state == WAITING {
		if ra.requestTime < req.Timestamp {
			deferReply = true
		} else if ra.requestTime == req.Timestamp && ra.nodeID < req.NodeID {
			deferReply = true
		}
	}

	if deferReply {
		ra.deferredReplies = append(ra.deferredReplies, req.NodeID)
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
		
		myID := ra.nodeID
		url := ra.nodes[req.NodeID]
		go func(tURL string) {
			payload := RAMessage{NodeID: myID, Timestamp: 0}
			data, _ := json.Marshal(payload)
			client := http.Client{Timeout: 1 * time.Second}
			client.Post(tURL+"/ra/reply", "application/json", bytes.NewBuffer(data))
		}(url)
	}
}

func (ra *RicartAgrawala) handleReply(w http.ResponseWriter, r *http.Request) {
	var req RAMessage
	json.NewDecoder(r.Body).Decode(&req)
	
	ra.addReply()
	w.WriteHeader(http.StatusOK)
}

func (ra *RicartAgrawala) addReply() {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	
	ra.replyCount++
	if ra.replyCount >= ra.expectedReplies {
		ra.cond.Broadcast()
	}
}

type RAStatus struct {
	NodeID          int    `json:"nodeId"`
	State           string `json:"state"`
	Clock           int    `json:"clock"`
	DeferredReplies []int  `json:"deferredReplies"`
	ReplyCount      int    `json:"replyCount"`
	ExpectedReplies int    `json:"expectedReplies"`
}

func (ra *RicartAgrawala) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Need to read values carefully without fully blocking if another process is in CRITICAL holding the Lock via sleep
	ra.mu.Lock()
	stateCopy := ra.state
	clockCopy := ra.clock.GetTime()
	replyCountCopy := ra.replyCount
	expectedRepliesCopy := ra.expectedReplies
	nodeIDCopy := ra.nodeID

	// Copy deferred replies
	deferredCopy := make([]int, len(ra.deferredReplies))
	copy(deferredCopy, ra.deferredReplies)
	ra.mu.Unlock()

	stateStr := "IDLE"
	if stateCopy == WAITING {
		stateStr = "WAITING"
	} else if stateCopy == CRITICAL_SECTION {
		stateStr = "CRITICAL_SECTION"
	}

	status := RAStatus{
		NodeID:          nodeIDCopy,
		State:           stateStr,
		Clock:           clockCopy,
		DeferredReplies: deferredCopy,
		ReplyCount:      replyCountCopy,
		ExpectedReplies: expectedRepliesCopy,
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
