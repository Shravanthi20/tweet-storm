package leader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"tweetstorm/algorithms"
	"tweetstorm/shared"
)

var initialWorkers = []string{
	"http://localhost:8001",
	"http://localhost:8002",
	"http://localhost:8003",
	"http://localhost:8004", // adding the fourth worker from simulate_crash configuration
}

var hashRing = algorithms.NewHashRing(500) // 500 virtual nodes for more even distribution
var leaderClock = algorithms.NewLamportClock("leader")
var eventMu sync.Mutex
var eventLogs []shared.EventLog

// State tracking for UI visualization
var stateMu sync.Mutex
var taskCounts = make(map[string]int)
var recentTasks = make(map[string]string)

func addEvent(node string, message string, timestamp int) {
	eventMu.Lock()
	defer eventMu.Unlock()

	eventLogs = append(eventLogs, shared.EventLog{
		Node:      node,
		Message:   message,
		Timestamp: timestamp,
	})
}

func forwardTask(tweet shared.Tweet) {
	ts := leaderClock.SendEvent()

	// Append a unique identifier to the tweet content for the hash key
	// so that identical tweet strings from the client get evenly distributed!
	hashKey := fmt.Sprintf("%s-%d-%d", tweet.Content, tweet.ID, time.Now().UnixNano())

	// Get the worker node URL via Consistent Hashing on the unique key
	workerURL := hashRing.GetNode(hashKey)

	if workerURL == "" {
		fmt.Println("[Leader] WARNING: No active workers in the hash ring! Task dropped.")
		return
	}

	addEvent("Leader", fmt.Sprintf("assigned task to %s", workerURL), ts)

	// Update UI State
	stateMu.Lock()
	taskCounts[workerURL]++
	recentTasks[workerURL] = tweet.Content
	stateMu.Unlock()

	task := shared.Task{
		Tweet:     tweet,
		Timestamp: ts,
	}

	data, _ := json.Marshal(task)

	// Post the task to the selected worker node's /task endpoint
<<<<<<< HEAD
	go func() {
		http.Post(workerURL+"/task", "application/json", bytes.NewBuffer(data))
	}()
=======
	http.Post(workerURL+"/task", "application/json", bytes.NewBuffer(data))
>>>>>>> b6801d1c89b95026b77de054393138270fcfca95
}

func startHealthChecks() {
	ticker := time.NewTicker(3 * time.Second)
	client := http.Client{Timeout: 1 * time.Second}

	for range ticker.C {
		for _, worker := range initialWorkers {
			// Ping the worker
			resp, err := client.Get(worker + "/worker/ping")
			if err != nil || resp.StatusCode != http.StatusOK {
				// Worker is unreachable, remove from ring
				hashRing.RemoveNode(worker)
				fmt.Printf("[Leader] Health check failed for %s. Removed from hash ring.\n", worker)
			} else {
				// Worker is reachable, add to ring
				hashRing.AddNode(worker)
			}

			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func HandleTweet(w http.ResponseWriter, r *http.Request) {

	var tweet shared.Tweet

	err := json.NewDecoder(r.Body).Decode(&tweet)
	if err != nil {
		http.Error(w, "Invalid tweet", http.StatusBadRequest)
		return
	}

	ts := leaderClock.Tick()
	addEvent("Leader", fmt.Sprintf("received tweet: %s", tweet.Content), ts)
	fmt.Println("Leader received tweet:", tweet.Content)

	forwardTask(tweet)

	w.Write([]byte("Tweet forwarded"))
}

func HandleEvents(w http.ResponseWriter, r *http.Request) {
	eventMu.Lock()
	defer eventMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(eventLogs)
}

// HandleState returns the current cluster state for frontend visualization
func HandleState(w http.ResponseWriter, r *http.Request) {
	stateMu.Lock()
	defer stateMu.Unlock()

	state := map[string]interface{}{
		"activeWorkers": hashRing.GetNodes(),
		"taskCounts":    taskCounts,
		"recentTasks":   recentTasks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(state)
}

<<<<<<< HEAD
// InitFailoverLeader is called by a Worker node when it wins a Bully election
func InitFailoverLeader() {
	// Initialize HashRing and Health checks
	for _, w := range initialWorkers {
		hashRing.AddNode(w)
	}
	go startHealthChecks()

	fmt.Println("[Leader Failover] Hash ring and health checks initialized.")
}

=======
>>>>>>> b6801d1c89b95026b77de054393138270fcfca95
func StartLeader(port string) {

	// Node 5 is the initial Leader
	algorithms.InitBully(5, port, 5)

<<<<<<< HEAD
	InitFailoverLeader()
=======
	// Initialize HashRing and Health checks
	for _, w := range initialWorkers {
		hashRing.AddNode(w)
	}
	go startHealthChecks()
>>>>>>> b6801d1c89b95026b77de054393138270fcfca95

	http.HandleFunc("/tweet", HandleTweet)
	http.HandleFunc("/events", HandleEvents)
	http.HandleFunc("/api/state", HandleState)

	fmt.Println("Leader running on port", port)

	// Will block until port error
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("[Leader Node 5] ERROR: Could not bind to port %s — %v\n", port, err)
		fmt.Println("[Leader Node 5] TIP: Kill any existing process on this port and retry.")
	}
}
