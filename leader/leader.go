package leader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"tweetstorm/shared"
)

var workers = []string{
	"http://localhost:8001/task",
	"http://localhost:8002/task",
	"http://localhost:8003/task",
}

var nextWorker = 0

func forwardTask(tweet shared.Tweet) {
	task := shared.Task{Tweet: tweet}

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

	fmt.Println("Leader received tweet:", tweet.Content)

	forwardTask(tweet)

	w.Write([]byte("Tweet forwarded"))
}

func StartLeader(port string) {

	http.HandleFunc("/tweet", handleTweet)

	fmt.Println("Leader running on port", port)

	http.ListenAndServe(":"+port, nil)
}
