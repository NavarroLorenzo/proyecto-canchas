package main

import (
	"log"
	"os"

	"search-api/config"
	"search-api/internal/cache"
	"search-api/internal/consumers"
	"search-api/internal/controllers"
	"search-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	cacheManager := cache.NewCacheManager(config.AppConfig)
	searchService := services.NewSearchService(cacheManager)
	searchController := controllers.NewSearchController(searchService)

	go func() {
		consumer, err := consumers.NewRabbitConsumer(os.Getenv("RABBITMQ_URL"), searchService)
		if err != nil {
			log.Fatalf("[RabbitMQ] Connection error: %v", err)
		}

		if err := consumer.Listen(os.Getenv("RABBITMQ_EXCHANGE"), "search_queue"); err != nil {
			log.Fatalf("[RabbitMQ] Listen error: %v", err)
		}
	}()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/search", searchController.Search)
	router.GET("/search2", searchController.Search)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("[HTTP] Search API running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
