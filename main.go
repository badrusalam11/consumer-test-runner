package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"consumer-test-runner/internal/config"
	"consumer-test-runner/internal/infrastructure/messaging"
)

// AutomationPayload defines the expected payload structure.
type AutomationPayload struct {
	TestSuiteID string `json:"testsuite_id"`
	Project     string `json:"project"`
}

func main() {
	// Load configuration from config.json.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to RabbitMQ using the dynamic configuration.
	conn, channel, err := messaging.ConnectToRabbitMQ(cfg.RabbitMQ.AMQPURL())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	// Declare the queue on the consumer channel to ensure it exists.
	_, err = channel.QueueDeclare(
		cfg.RabbitMQ.QueueName, // name
		true,                   // durable
		false,                  // autoDelete
		false,                  // exclusive
		false,                  // noWait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Create a separate channel for publishing messages.
	// pubChannel, err := conn.Channel()
	// if err != nil {
	// 	log.Fatalf("Failed to create publishing channel: %v", err)
	// }
	// defer pubChannel.Close()

	// Create a publisher using the queue name from config.
	publisher := messaging.NewRabbitMQPublisher(channel, cfg.RabbitMQ.QueueName)

	// Set up a consumer on the same queue with manual acks.
	msgs, err := channel.Consume(
		cfg.RabbitMQ.QueueName, // queue
		"",                     // consumer tag – let RabbitMQ generate one
		false,                  // autoAck is false; we'll acknowledge manually
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// Launch a goroutine to process incoming messages.
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// Unmarshal the message into AutomationPayload.
			var payload AutomationPayload
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				// Acknowledge the message even if there's an unmarshaling error.
				d.Ack(false)
				continue
			}

			// Marshal the payload to prepare the HTTP request.
			reqBody, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error marshaling payload: %v", err)
				d.Ack(false)
				continue
			}

			// Hit the endpoint.
			resp, err := http.Post("http://localhost:6000/automation/run", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				log.Printf("HTTP request error: %v", err)
				// Reproduce the message to RabbitMQ.
				republishMessage(publisher, d.Body)
				d.Ack(false)
				continue
			}
			// Close response body.
			resp.Body.Close()

			// Check if the status code is 200.
			if resp.StatusCode != http.StatusOK {
				log.Printf("HTTP response status not OK: %d", resp.StatusCode)
				// Republish the message if not 200.
				republishMessage(publisher, d.Body)
			} else {
				log.Printf("Message processed successfully")
			}

			// Acknowledge the original message.
			d.Ack(false)
		}
	}()

	log.Printf("Consumer is running. Waiting for messages...")
	// Block forever.
	select {}
}

// republishMessage publishes the message back to RabbitMQ after a short delay.
func republishMessage(publisher messaging.Publisher, body []byte) {
	// Optional delay to avoid tight loops.
	time.Sleep(2 * time.Second)
	if err := publisher.Publish(body); err != nil {
		log.Printf("Failed to republish message: %v", err)
	} else {
		log.Printf("Message reproduced to RabbitMQ")
	}
}
