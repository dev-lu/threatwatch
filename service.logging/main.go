package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {

	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", "auth_topic", 0)
	if err != nil {
		fmt.Println("Error to connect to Kafka")
	}
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))

	http.HandleFunc("/api/v1/logging/health", healthcheckHandler)
	log.Fatal(http.ListenAndServe(":4000", nil))

	batch := conn.ReadBatch(1e3, 1e9) // Min message size 1kb, max 1GB
	bytes := make([]byte, 1e3)
	for {
		_, err := batch.Read(bytes)
		if err != nil {
			fmt.Println("could not read: ", bytes)
		}
		fmt.Println((string(bytes)))
	}

}
