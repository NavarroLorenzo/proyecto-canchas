package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	SolrURL              string
	SolrCore             string
	RabbitMQURL          string
	RabbitMQExchange     string
	RabbitMQQueue        string
	MemcachedURL         string
	LocalCacheSize       int
	LocalCacheTTLMinutes int
	MemcachedTTLMinutes  int
	CanchasAPIURL        string
}

var AppConfig *Config

// LoadConfig carga las variables de entorno
func LoadConfig() {
	// Cargar .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	localCacheSize, _ := strconv.Atoi(getEnv("LOCAL_CACHE_SIZE", "1000"))
	localCacheTTL, _ := strconv.Atoi(getEnv("LOCAL_CACHE_TTL_MINUTES", "5"))
	memcachedTTL, _ := strconv.Atoi(getEnv("MEMCACHED_TTL_MINUTES", "10"))

	AppConfig = &Config{
		Port:                 getEnv("PORT", "8083"),
		SolrURL:              getEnv("SOLR_URL", "http://localhost:8983/solr"),
		SolrCore:             getEnv("SOLR_CORE", "canchas"),
		RabbitMQURL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange:     getEnv("RABBITMQ_EXCHANGE", "canchas_events"),
		RabbitMQQueue:        getEnv("RABBITMQ_QUEUE", "canchas_queue"),
		MemcachedURL:         getEnv("MEMCACHED_URL", "localhost:11211"),
		LocalCacheSize:       localCacheSize,
		LocalCacheTTLMinutes: localCacheTTL,
		MemcachedTTLMinutes:  memcachedTTL,
		CanchasAPIURL:        getEnv("CANCHAS_API_URL", "http://localhost:8081"),
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
