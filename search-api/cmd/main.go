package main

import (
	"log"
	"os"
	"time"

	"search-api/config"
	"search-api/internal/cache"
	"search-api/internal/consumers"
	"search-api/internal/controllers"
	"search-api/internal/repositories"
	"search-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	cacheManager := cache.NewCacheManager(config.AppConfig)
	solrRepo := repositories.NewSolrRepository(config.AppConfig.SolrURL, config.AppConfig.SolrCore, nil)
	searchService := services.NewSearchService(cacheManager, solrRepo, config.AppConfig.CanchasAPIURL)
	searchController := controllers.NewSearchController(searchService)

	go func() {
		const maxAttempts = 5
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if err := searchService.ReindexAllCanchas(); err != nil {
				log.Printf("[Reindex] Attempt %d/%d failed: %v", attempt, maxAttempts, err)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Printf("[Reindex] Completed successfully on attempt %d", attempt)
			return
		}
		log.Printf("[Reindex] Failed after %d attempts, continuing without a fresh index", maxAttempts)
	}()

	go func() {
		consumer, err := consumers.NewRabbitConsumer(config.AppConfig.RabbitMQURL, searchService)
		if err != nil {
			log.Fatalf("[RabbitMQ] Connection error: %v", err)
		}

		if err := consumer.Listen(config.AppConfig.RabbitMQExchange, config.AppConfig.RabbitMQQueue); err != nil {
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
