package main

import (
	"canchas-api/config"
	"canchas-api/internal/clients"
	"canchas-api/internal/controllers"
	"canchas-api/internal/messaging"
	"canchas-api/internal/repositories"
	"canchas-api/internal/services"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	config.LoadConfig()
	db := connectMongoDB()

	publisher, err := messaging.NewRabbitMQPublisher()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	reservaClient := clients.NewReservaClient()

	canchaRepo := repositories.NewCanchaRepository(db)
	canchaService := services.NewCanchaService(canchaRepo, publisher, reservaClient)
	canchaController := controllers.NewCanchaController(canchaService)

	router := setupRouter(canchaController)

	port := config.AppConfig.Port
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func connectMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoURI)

	var client *mongo.Client
	var err error

	for i := 0; i < 30; i++ {
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			err = client.Ping(ctx, nil)
			if err == nil {
				log.Println("MongoDB connection established successfully")
				return client.Database(config.AppConfig.MongoDatabase)
			}
		}

		log.Printf("Failed to connect to MongoDB (attempt %d/30): %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("Could not connect to MongoDB after 30 attempts: %v", err)
	return nil
}

func setupRouter(canchaController *controllers.CanchaController) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "canchas-api",
		})
	})

	// Rutas públicas (cualquiera puede ver canchas)
	router.GET("/canchas", canchaController.GetAll)
	router.GET("/canchas/:id", canchaController.GetByID)

	// Rutas protegidas (SOLO ADMIN puede crear/editar/eliminar)
	// Por ahora sin middleware, pero deberías añadirlo
	router.POST("/canchas", canchaController.Create)       // TODO: Añadir AdminMiddleware
	router.PUT("/canchas/:id", canchaController.Update)    // TODO: Añadir AdminMiddleware
	router.DELETE("/canchas/:id", canchaController.Delete) // TODO: Añadir AdminMiddleware

	log.Println("Routes configured successfully")
	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
