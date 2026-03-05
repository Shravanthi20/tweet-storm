package shared

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func InitDB(uri string, dbName string) {

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	DB = client.Database(dbName)
}

type Tweet struct {
	ID      int    `json:"id" bson:"id"`
	Content string `json:"content" bson:"content"`
}

type Task struct {
	Tweet     Tweet `json:"tweet"`
	Timestamp int   `json:"timestamp"`
}

type Result struct {
	WordCounts map[string]int `json:"wordCounts"`
}

type EventLog struct {
	Node      string `json:"node"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
}

type WorkerInfo struct {
	ID   string `json:"id"`
	Port string `json:"port"`
}
