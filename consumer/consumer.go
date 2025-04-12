package consumer

import (
	"consumer-test-runner/config"
	"consumer-test-runner/handler"
	"consumer-test-runner/model"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

var req model.AutomationRequest

func StartConsumer() {
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.AppConfig.RabbitMQUser,
		config.AppConfig.RabbitMQPass,
		config.AppConfig.RabbitMQHost,
		config.AppConfig.RabbitMQPort,
	)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	exchange := config.AppConfig.RabbitMQExchange
	if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		log.Fatalf("Failed to declare exchange: %s", err)
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	if err := ch.QueueBind(q.Name, "", exchange, false, nil); err != nil {
		log.Fatalf("Failed to bind queue: %s", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %s", err)
	}

	fmt.Println("[*] Waiting for messages. To exit press CTRL+C")
	for d := range msgs {
		fmt.Printf("[x] Received: %s\n", d.Body)
		if err := json.Unmarshal(d.Body, &req); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}
		if err := handler.HandleRequest(req); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}
