package messaging

import (
	"github.com/streadway/amqp"
)

// ConnectToRabbitMQ connects to the RabbitMQ server using the provided URL and returns the connection and channel.
func ConnectToRabbitMQ(amqpURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

// Publisher defines the interface for a message publisher.
type Publisher interface {
	Publish(message []byte) error
}

// RabbitMQPublisher is a concrete implementation that publishes messages to RabbitMQ.
type RabbitMQPublisher struct {
	Channel   *amqp.Channel
	QueueName string
}

// NewRabbitMQPublisher creates a new RabbitMQPublisher.
func NewRabbitMQPublisher(channel *amqp.Channel, queueName string) *RabbitMQPublisher {
	_, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		panic(err) // handle error appropriately in production code
	}
	return &RabbitMQPublisher{
		Channel:   channel,
		QueueName: queueName,
	}
}

// Publish sends the message to RabbitMQ.
func (p *RabbitMQPublisher) Publish(message []byte) error {
	return p.Channel.Publish(
		"",          // exchange
		p.QueueName, // routing key (queue name)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
}
