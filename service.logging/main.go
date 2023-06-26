package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"service.logging/db"
)

const (
	brokerAddress = "kafka:9092"
	topic         = "auth_topic"
	groupID       = "auth_consumer_group"
)

func main() {
	var r *kafka.Reader
	var err error

	for {
		r, err = createKafkaReader()
		if err == nil {
			break
		}

		log.Println("Failed to connect to Kafka. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	// Create context for handling graceful shutdown of consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channel to receive OS interrupt signals (e.g., Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Start goroutine to handle OS interrupt signals
	go func() {
		<-sigChan
		log.Println("Received OS interrupt signal. Shutting down...")
		cancel()
	}()

	// Start consuming messages from the Kafka topic
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}

		// Print message timestamp and value and save to DB
		fmt.Printf("[%s] %s\n", msg.Time, string(msg.Value))
		if err := addToDb(msg.Time, string(msg.Value)); err != nil {
			log.Println("Error adding log to DB:", err)
		}
	}
}

func createKafkaReader() (*kafka.Reader, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokerAddress, ","),
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        time.Second,
		CommitInterval: time.Second,
	})

	// Dummy ReadMessage to ensure connection is established
	_, err := r.ReadMessage(context.Background())
	if err != nil {
		r.Close() // Close reader if the connection unsuccessful
		return nil, err
	}

	return r, nil
}

func addToDb(timestamp time.Time, message string) error {
	session, err := db.GetSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Insert the log into the Cassandra table
	if err := session.Query(
		"INSERT INTO logging.auth_log (timestamp, message) VALUES (?, ?)",
		timestamp,
		message,
	).Exec(); err != nil {
		return err
	}

	log.Println("Added log to DB")
	return nil
}
