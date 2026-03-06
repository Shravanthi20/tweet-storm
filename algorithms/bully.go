package algorithms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	nodeID   int
	leaderID int
	nodePort string
	nodes    map[int]string
	bullyMu       sync.Mutex
	electionMutex sync.Mutex
	inElection    bool
)

func InitBully(id int, port string, initialLeader int) {
	nodeID = id
	nodePort = port
	leaderID = initialLeader

	leaderIP := os.Getenv("LEADER_IP")
	if leaderIP == "" {
		leaderIP = "localhost"
	}
	workerIP := os.Getenv("WORKER_IP")
	if workerIP == "" {
		workerIP = "localhost"
	}

	nodes = map[int]string{
		1: "http://" + workerIP + ":8001",
		2: "http://" + workerIP + ":8002",
		3: "http://" + workerIP + ":8003",
		4: "http://" + workerIP + ":8004",
		5: "http://" + leaderIP + ":8000",
	}

	// Register Bully HTTP Handlers
	http.HandleFunc("/bully/election", handleReceiveElectionMessage)
	http.HandleFunc("/bully/coordinator", handleReceiveCoordinator)
	http.HandleFunc("/bully/ping", handlePing)
	http.HandleFunc("/bully/status", handleBullyStatus)

	// Start pinging the leader to detect failures
	if nodeID != leaderID {
		go StartPing()
	}
}

func StartPing() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	consecutiveFailures := 0

	for range ticker.C {
		currentLeader := GetLeader()
		if currentLeader == nodeID {
			continue // I am the leader, no need to ping myself
		}

		leaderURL := nodes[currentLeader]
		if leaderURL == "" {
			continue
		}

		client := http.Client{Timeout: 1 * time.Second}
		_, err := client.Get(leaderURL + "/bully/ping")
		if err != nil {
			consecutiveFailures++
			if consecutiveFailures >= 2 {
				fmt.Printf("[Node %d] Leader failure detected\n", nodeID)
				StartElection(nodeID)
				consecutiveFailures = 0 // Reset after starting election
			}
		} else {
			consecutiveFailures = 0 // Reset on successful ping
		}
	}
}

func StartElection(id int) {
	electionMutex.Lock()
	if inElection {
		electionMutex.Unlock()
		return
	}
	inElection = true
	electionMutex.Unlock()

	defer func() {
		electionMutex.Lock()
		inElection = false
		electionMutex.Unlock()
	}()

	fmt.Printf("[Node %d] Starting election\n", id)

	higherNodesResponded := false

	// Send an election message to all nodes with a higher ID
	for nID, url := range nodes {
		if nID > id {
			success := SendElectionMessage(nID, url)
			if success {
				higherNodesResponded = true
			}
		}
	}

	// If no higher node responded, this node becomes the leader
	if !higherNodesResponded {
		becomeLeader()
	}
}

func SendElectionMessage(targetID int, targetURL string) bool {
	client := http.Client{Timeout: 2 * time.Second}

	payload := map[string]int{"senderID": nodeID}
	data, _ := json.Marshal(payload)

	resp, err := client.Post(targetURL+"/bully/election", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func handleReceiveElectionMessage(w http.ResponseWriter, r *http.Request) {
	var payload map[string]int
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	senderID := payload["senderID"]
	ReceiveElectionMessage(senderID)

	// Acknowledge the election message
	w.WriteHeader(http.StatusOK)
}

func ReceiveElectionMessage(senderID int) {
	fmt.Printf("[Node %d] Responding to election\n", nodeID)
	// Start an election to see if we can be leader
	go StartElection(nodeID)
}

// OnBecomeLeader is an optional callback that fires when this node wins an election
var OnBecomeLeader func()

func becomeLeader() {
	bullyMu.Lock()
	leaderID = nodeID
	bullyMu.Unlock()

	fmt.Printf("[Node %d] Broadcasting coordinator message\n", nodeID)
	SendCoordinator(nodeID)

	// Start taking leader traffic on port 8000 for clients/frontend
	go func() {
		// Only try to handle traffic on 8000 if we are not the initial leader (who is already there)
		if nodePort != "8000" {
			// Trigger any application-level routing setup (e.g. hashring)
			if OnBecomeLeader != nil {
				OnBecomeLeader()
			}

			// Custom mux for the leader endpoints to avoid path collision
			// with the worker endpoints on the default mux
			leaderMux := http.NewServeMux()

			// This expects the leader package functions handleTweet / handleEvents are exposed.
			// We will register a forwarding function that directs client traffic.
			leaderMux.HandleFunc("/tweet", ForwardToCurrentNodeHandler("/tweet"))
			leaderMux.HandleFunc("/events", ForwardToCurrentNodeHandler("/events"))
			leaderMux.HandleFunc("/api/state", ForwardToCurrentNodeHandler("/api/state"))
			leaderMux.HandleFunc("/bully/status", ForwardToCurrentNodeHandler("/bully/status"))

			// Try to bind to port 8000, retrying if it's still in use (TIME_WAIT from crashed node)
			for {
				fmt.Printf("[Node %d] Taking over client traffic on port 8000\n", nodeID)
				err := http.ListenAndServe(":8000", leaderMux)
				if err != nil {
					fmt.Printf("[Node %d] Port 8000 still in use, retrying in 2 seconds...\n", nodeID)
					time.Sleep(2 * time.Second)
				} else {
					break // ListenAndServe blocks on success, so we only break if it returns unexpectedly without error
				}
			}
		}
	}()

	fmt.Printf("\n[Node %d] New leader elected\n", nodeID)
}

// Helper to proxy requests directed at port 8000 to the current node's port
func ForwardToCurrentNodeHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := "http://localhost:" + nodePort + path

		req, _ := http.NewRequest(r.Method, targetURL, r.Body)
		req.Header = r.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error forwarding", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		w.Write(buf.Bytes())
	}
}

func SendCoordinator(id int) {
	for nID, url := range nodes {
		if nID != id {
			client := http.Client{Timeout: 1 * time.Second}
			payload := map[string]int{"leaderID": id}
			data, _ := json.Marshal(payload)

			_, _ = client.Post(url+"/bully/coordinator", "application/json", bytes.NewBuffer(data))
		}
	}
}

func handleReceiveCoordinator(w http.ResponseWriter, r *http.Request) {
	var payload map[string]int
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ReceiveCoordinator(payload["leaderID"])
	w.WriteHeader(http.StatusOK)
}

func ReceiveCoordinator(newLeaderID int) {
	bullyMu.Lock()
	leaderID = newLeaderID
	bullyMu.Unlock()
	// Optionally log here, but user asked for specific logs which are handled.
}

func GetLeader() int {
	bullyMu.Lock()
	defer bullyMu.Unlock()
	return leaderID
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleBullyStatus(w http.ResponseWriter, r *http.Request) {
	// CORS headers for the frontend visualization
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	bullyMu.Lock()
	currentLeader := leaderID
	currentNode := nodeID
	bullyMu.Unlock()

	electionMutex.Lock()
	currentElectionState := inElection
	electionMutex.Unlock()

	status := map[string]interface{}{
		"nodeId":     currentNode,
		"leaderId":   currentLeader,
		"inElection": currentElectionState,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
