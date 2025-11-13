package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	MongoURI         string
	MongoDatabase    string
	RabbitMQURL      string
	RabbitMQExchange string
	RabbitMQQueue    string
	UsersAPIURL      string
	ReservasAPIURL   string
}

var AppConfig *Config

// LoadConfig carga las variables de entorno
func LoadConfig() {
	// Cargar .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		Port:             getEnv("PORT", "8081"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:    getEnv("MONGO_DATABASE", "canchas_db"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "canchas_events"),
		RabbitMQQueue:    getEnv("RABBITMQ_QUEUE", "canchas_queue"),
		UsersAPIURL:      getEnv("USERS_API_URL", "http://localhost:8080"),
		ReservasAPIURL:   getEnv("RESERVAS_API_URL", "http://localhost:8082"),
	}

	log.Println("Configuration loaded successfully")
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
