package main

import (
	"log"
	"os"

	"search-api/config"
	"search-api/internal/consumers"
	"search-api/internal/controllers"
	"search-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1️⃣ Cargar configuración desde .env
	config.LoadConfig()

	// 2️⃣ Inicializar servicio y controlador
	searchService := services.NewSearchService()
	searchController := controllers.NewSearchController(searchService)

	// 3️⃣ Iniciar consumer de RabbitMQ en segundo plano
	go func() {
		consumer, err := consumers.NewRabbitConsumer(os.Getenv("RABBITMQ_URL"), searchService)
		if err != nil {
			log.Fatalf("[RabbitMQ] Connection error: %v", err)
		}

		err = consumer.Listen(os.Getenv("RABBITMQ_EXCHANGE"), "search_queue")
		if err != nil {
			log.Fatalf("[RabbitMQ] Listen error: %v", err)
		}
	}()

	// 4️⃣ Iniciar servidor HTTP
	router := gin.Default()

	// Endpoint principal /search usa el controlador para soportar filtros avanzados
	router.GET("/search", searchController.Search)

	// También podés montar el controller completo (alias)
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
