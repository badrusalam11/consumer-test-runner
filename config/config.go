package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQHost     string
	RabbitMQPort     int
	RabbitMQUser     string
	RabbitMQPass     string
	RabbitMQExchange string
	AutomationAPI    string
}

var AppConfig Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = Config{
		RabbitMQHost:     viper.GetString("RABBITMQ_HOST"),
		RabbitMQPort:     viper.GetInt("RABBITMQ_PORT"),
		RabbitMQUser:     viper.GetString("RABBITMQ_USER"),
		RabbitMQPass:     viper.GetString("RABBITMQ_PASS"),
		RabbitMQExchange: viper.GetString("RABBITMQ_EXCHANGE"),
		AutomationAPI:    viper.GetString("AUTOMATION_API"),
	}
}
