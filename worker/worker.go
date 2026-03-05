package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"tweetstorm/shared"
)

var (
	globalWordCount = make(map[string]int)
	mu              sync.Mutex
)

func processTweet(tweet string) {
	words := strings.Fields(tweet)

	mu.Lock()
	defer mu.Unlock()

	for _, w := range words {
		globalWordCount[w]++
	}
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	var task shared.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid task", http.StatusBadRequest)
		return
	}

	processTweet(task.Tweet.Content)

	fmt.Println("Processed:", task.Tweet.Content)

	json.NewEncoder(w).Encode(globalWordCount)
}

func StartWorker(port string) {
	http.HandleFunc("/task", handleTask)

	fmt.Println("Worker running on port", port)

	http.ListenAndServe(":"+port, nil)
}
