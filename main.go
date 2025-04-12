package main

import (
	"consumer-test-runner/config"
	"consumer-test-runner/consumer"
)

func main() {
	config.LoadConfig()
	consumer.StartConsumer()
}
