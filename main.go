package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"tweetstorm/client"
	"tweetstorm/leader"
	"tweetstorm/worker"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

func main() {

	loadEnv()

	role := flag.String("role", "leader", "leader | worker | client")
	port := flag.String("port", "8000", "port number")

	flag.Parse()

	switch *role {

	case "leader":
		leader.StartLeader(*port)

	case "worker":
		worker.StartWorker(*port)

	case "client":
		client.StartClient()

	default:
		fmt.Println("Unknown role")
	}
}
