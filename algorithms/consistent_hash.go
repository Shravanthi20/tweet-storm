package algorithms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Node struct {
	ID   int
	Port string
}

var (
	myID       int
	myPort     string
	leaderPort string
	nodes      []Node
)

func InitBully(id int, port string, allNodes []Node) {
	myID = id
	myPort = port
	nodes = allNodes
}

func StartElection() {
	fmt.Println("Starting election...")

	// Send election message to all higher ID nodes
	for _, n := range nodes {
		if n.ID > myID {
			go sendElection(n)
		}
	}

	// Wait for timeout to see if anyone responds
	time.Sleep(1 * time.Second)

	// If no response, I am the leader
	if leaderPort == "" {
		leaderPort = myPort
		fmt.Println("I am the leader!")
		notifyLeader()
	}
}

func sendElection(n Node) {
	url := fmt.Sprintf("http://localhost:%s/election", n.Port)
	payload := map[string]int{"id": myID}

	jsonBody, _ := json.Marshal(payload)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		fmt.Println("Node down:", n.Port)
	}
}

func handleElection(w http.ResponseWriter, r *http.Request) {
	var msg map[string]int
	json.NewDecoder(r.Body).Decode(&msg)

	fmt.Println("Election message from:", msg["id"])

	// Send OK back
	w.WriteHeader(http.StatusOK)

	// Start my own election
	go StartElection()
}

func notifyLeader() {
	for _, n := range nodes {
		if n.Port != myPort {
			go sendLeader(n)
		}
	}
}

func sendLeader(n Node) {
	url := fmt.Sprintf("http://localhost:%s/leader", n.Port)
	payload := map[string]string{"port": leaderPort}

	jsonBody, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

func handleLeader(w http.ResponseWriter, r *http.Request) {
	var msg map[string]string
	json.NewDecoder(r.Body).Decode(&msg)
	leaderPort = msg["port"]
	fmt.Println("New leader:", leaderPort)
}
