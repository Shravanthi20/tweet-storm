package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tweetstorm/shared"
)

var tweets = []string{
	"distributed systems are powerful",
	"apache storm processes streams",
	"go is great for concurrency",
	"real time processing is fun",
}

func StartClient() {

	i := 0

	for {

		tweet := shared.Tweet{
			ID:      i,
			Content: tweets[i%len(tweets)],
		}

		data, _ := json.Marshal(tweet)

		http.Post(
			"http://localhost:8000/tweet",
			"application/json",
			bytes.NewBuffer(data),
		)

		fmt.Println("Client sent:", tweet.Content)

		i++

		time.Sleep(2 * time.Second)
	}
}
