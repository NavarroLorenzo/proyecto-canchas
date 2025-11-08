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
	// Cargar configuración
	config.LoadConfig()

	// Conectar a MongoDB
	db := connectMongoDB()

	// Conectar a RabbitMQ
	publisher, err := messaging.NewRabbitMQPublisher()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	// Inicializar clientes
	userClient := clients.NewUserClient()

	// Inicializar repositorios
	canchaRepo := repositories.NewCanchaRepository(db)

	// Inicializar servicios
	canchaService := services.NewCanchaService(canchaRepo, userClient, publisher)

	// Inicializar controladores
	canchaController := controllers.NewCanchaController(canchaService)

	// Configurar Gin
	router := setupRouter(canchaController)

	// Iniciar servidor
	port := config.AppConfig.Port
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// connectMongoDB establece la conexión con MongoDB
func connectMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoURI)

	// Reintentar conexión hasta 30 veces
	var client *mongo.Client
	var err error

	for i := 0; i < 30; i++ {
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			// Verificar la conexión
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

// setupRouter configura las rutas de la API
func setupRouter(canchaController *controllers.CanchaController) *gin.Engine {
	router := gin.Default()

	// Middleware CORS
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "canchas-api",
		})
	})

	// Rutas de canchas
	canchas := router.Group("/canchas")
	{
		canchas.POST("", canchaController.Create)
		canchas.GET("", canchaController.GetAll)
		canchas.GET("/:id", canchaController.GetByID)
		canchas.PUT("/:id", canchaController.Update)
		canchas.DELETE("/:id", canchaController.Delete)
		canchas.GET("/owner/:owner_id", canchaController.GetByOwnerID)
	}

	log.Println("Routes configured successfully")
	return router
}

// corsMiddleware configura CORS para desarrollo
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
