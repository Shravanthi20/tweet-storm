package leader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"tweetstorm/algorithms"
	"tweetstorm/shared"

	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getInitialWorkers() []string {
	workerIP := os.Getenv("WORKER_IP")
	if workerIP == "" {
		workerIP = "localhost"
	}
	return []string{
		"http://" + workerIP + ":8001",
		"http://" + workerIP + ":8002",
		"http://" + workerIP + ":8003",
		"http://" + workerIP + ":8004", // adding the fourth worker from simulate_crash configuration
	}
}

var mongoClient *mongo.Client
var tweetCollection *mongo.Collection

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
	go func() {
		http.Post(workerURL+"/task", "application/json", bytes.NewBuffer(data))
	}()
}

func startHealthChecks() {
	ticker := time.NewTicker(3 * time.Second)
	client := http.Client{Timeout: 1 * time.Second}

	for range ticker.C {
		for _, worker := range getInitialWorkers() {
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
	// CORS headers for the frontend React app to post directly
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var tweet shared.Tweet

	err := json.NewDecoder(r.Body).Decode(&tweet)
	if err != nil {
		fmt.Println("[Leader] ERROR: Failed to decode tweet JSON:", err)
		http.Error(w, "Invalid tweet format: "+err.Error(), http.StatusBadRequest)
		return
	}

	ts := leaderClock.Tick()
	addEvent("Leader", fmt.Sprintf("received tweet: %s", tweet.Content), ts)
	fmt.Println("Leader received tweet STARTING MONGO SAVE:", tweet.Content)

	// Save to MongoDB if available
	if tweetCollection != nil {
		fmt.Println("[Leader DB] Attempting to save tweet to Mongo...")
		tweet.ID = int(time.Now().UnixMilli()) // ensure unique ID for manual tweets
		res, err := tweetCollection.InsertOne(context.TODO(), tweet)
		if err != nil {
			fmt.Printf("[Leader DB] ERROR SAVING TWEET: %v\n", err)
		} else {
			fmt.Println("[Leader DB] Tweet saved to MongoDB! ID:", res.InsertedID)
		}
	} else {
		fmt.Println("[Leader DB] FATAL: tweetCollection is nil! MongoDB was never initialized.")
	}

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

// InitFailoverLeader is called by a Worker node when it wins a Bully election
func InitFailoverLeader() {
	// Initialize HashRing and Health checks
	for _, w := range getInitialWorkers() {
		hashRing.AddNode(w)
	}
	go startHealthChecks()

	fmt.Println("[Leader Failover] Hash ring and health checks initialized.")
}

func StartLeader(port string) {

	// Initialize MongoDB Connection
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("[Leader DB] WARNING: Could not connect to MongoDB:", err)
	} else {
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			fmt.Println("[Leader DB] WARNING: MongoDB ping failed:", err)
		} else {
			mongoClient = client
			tweetCollection = client.Database("tweetstorm").Collection("tweets")
			fmt.Println("[Leader DB] Successfully connected to MongoDB!")
		}
	}

	// Node 5 is the initial Leader
	algorithms.InitBully(5, port, 5)

	InitFailoverLeader()

	http.HandleFunc("/tweet", HandleTweet)
	http.HandleFunc("/events", HandleEvents)
	http.HandleFunc("/api/state", HandleState)

	fmt.Println("Leader running on port", port)

	// Will block until port error
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("[Leader Node 5] ERROR: Could not bind to port %s — %v\n", port, err)
		fmt.Println("[Leader Node 5] TIP: Kill any existing process on this port and retry.")
	}
}
