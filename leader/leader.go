package leader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"tweetstorm/algorithms"
	"tweetstorm/shared"
)

var workers = []string{
	"http://localhost:8001/task",
	"http://localhost:8002/task",
	"http://localhost:8003/task",
}

var nextWorker = 0
var leaderClock = algorithms.NewLamportClock("leader")
var eventMu sync.Mutex
var eventLogs []shared.EventLog

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
	addEvent("Leader", "assigned task to worker", ts)
	task := shared.Task{
		Tweet:     tweet,
		Timestamp: ts,
	}

	data, _ := json.Marshal(task)

	workerURL := workers[nextWorker]

	nextWorker = (nextWorker + 1) % len(workers)

	http.Post(workerURL, "application/json", bytes.NewBuffer(data))
}

func handleTweet(w http.ResponseWriter, r *http.Request) {

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

func handleEvents(w http.ResponseWriter, r *http.Request) {
	eventMu.Lock()
	defer eventMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(eventLogs)
}

func StartLeader(port string) {

	http.HandleFunc("/tweet", handleTweet)
	http.HandleFunc("/events", handleEvents)

	fmt.Println("Leader running on port", port)

	http.ListenAndServe(":"+port, nil)
}
