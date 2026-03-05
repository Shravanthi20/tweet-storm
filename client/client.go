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
		// Send a burst of 4 tweets concurrently (or quickly)
		for j := 0; j < 4; j++ {
			tweet := shared.Tweet{
				ID:      i,
				Content: tweets[i%len(tweets)],
			}

			data, _ := json.Marshal(tweet)

			go func(t shared.Tweet, d []byte) {
				http.Post(
					"http://localhost:8000/tweet",
					"application/json",
					bytes.NewBuffer(d),
				)
				fmt.Println("Client sent:", t.Content)
			}(tweet, data)

			i++
		}

		time.Sleep(5 * time.Second)
	}
}
