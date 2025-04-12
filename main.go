package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

// AutomationRequest represents the incoming message format
type AutomationRequest struct {
	Project     string `json:"project"`
	TestsuiteID string `json:"testsuite_id"`
	TotalSteps  int    `json:"total_steps"`
	Retries     int    `json:"retries"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	exchangeName := "automation"
	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %s", err)
	}

	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	err = ch.QueueBind(q.Name, "", exchangeName, false, nil)
	if err != nil {
		log.Fatalf("Failed to bind queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %s", err)
	}

	fmt.Println("[*] Waiting for messages. To exit press CTRL+C")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("[x] Received: %s\n", d.Body)

			var req AutomationRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			err := handleRequest(req, ch, exchangeName, d.Body)
			if err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()

	<-forever
}

func handleRequest(data AutomationRequest, ch *amqp.Channel, exchangeName string, originalBody []byte) error {
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:6000/automation/run", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("[✓] Response : %s\n", respBody)

	return nil
}
