package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"tweetstorm/algorithms"
	"tweetstorm/leader"
	"tweetstorm/shared"
)

var (
	globalWordCount = make(map[string]int)
	raNode          *algorithms.RicartAgrawala
	taskMu          sync.Mutex
)

func processTweet(tweet string) {
	taskMu.Lock()
	defer taskMu.Unlock()

	words := strings.Fields(tweet)

	if raNode != nil {
		raNode.RequestLock()
	}

	for _, w := range words {
		globalWordCount[w]++
	}

	// Artificial delay to visualize the Lock in the UI
	time.Sleep(2 * time.Second)

	if raNode != nil {
		raNode.ReleaseLock()
	}
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	var task shared.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid task", http.StatusBadRequest)
		return
	}

	// Update Lamport clock on the worker based on task timestamp
	if raNode != nil {
		// Use a dummy method or just acknowledge we received an event timestamp
		// Depending on where LamportClock is exported.
		// For simplicity, RA node's handleRequest manages its own clock.
	}

	processTweet(task.Tweet.Content)

	fmt.Println("Processed:", task.Tweet.Content)

	json.NewEncoder(w).Encode(globalWordCount)
}

func StartWorker(port string) {
	nodeID := 0
	switch port {
	case "8001":
		nodeID = 1
	case "8002":
		nodeID = 2
	case "8003":
		nodeID = 3
	case "8004":
		nodeID = 4
	}

	workerIP := os.Getenv("WORKER_IP")
	if workerIP == "" {
		workerIP = "localhost"
	}

	workerNodes := map[int]string{
		1: "http://" + workerIP + ":8001",
		2: "http://" + workerIP + ":8002",
		3: "http://" + workerIP + ":8003",
		4: "http://" + workerIP + ":8004",
	}

	if nodeID != 0 {
		// Setup Bully Failover Callback so this worker can act as the new leader
		algorithms.OnBecomeLeader = func() {
			leader.InitFailoverLeader()
		}

		// Initialize Bully on worker (Node 5 is initial leader)
		algorithms.InitBully(nodeID, port, 5)

		// Initialize Ricart-Agrawala Mutual Exclusion
		lamportClock := algorithms.NewLamportClock(fmt.Sprintf("worker-%d", nodeID))
		raNode = algorithms.NewRicartAgrawala(nodeID, port, lamportClock, workerNodes)
	}

	http.HandleFunc("/task", handleTask)
	http.HandleFunc("/worker/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Worker running on port", port)

	http.ListenAndServe(":"+port, nil)
}
