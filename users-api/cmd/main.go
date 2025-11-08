package main

import (
	"log"
	"time"
	"users-api/config"
	"users-api/internal/controllers"
	"users-api/internal/domain"
	"users-api/internal/middleware"
	"users-api/internal/repositories"
	"users-api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Cargar configuración
	config.LoadConfig()

	// Conectar a la base de datos
	db := connectDatabase()

	// Migrar modelos
	migrateDatabase(db)

	// Inicializar repositorios
	userRepo := repositories.NewUserRepository(db)

	// Inicializar servicios
	userService := services.NewUserService(userRepo)

	// Inicializar controladores
	userController := controllers.NewUserController(userService)

	// Configurar Gin
	router := setupRouter(userController)

	// Iniciar servidor
	port := config.AppConfig.Port
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// connectDatabase establece la conexión con MySQL
func connectDatabase() *gorm.DB {
	dsn := config.AppConfig.GetDSN()

	var db *gorm.DB
	var err error

	// Reintentar conexión hasta 30 veces (30 segundos)
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			log.Println("Database connection established successfully")

			// Configurar pool de conexiones
			sqlDB, err := db.DB()
			if err != nil {
				log.Fatalf("Failed to get database instance: %v", err)
			}

			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetConnMaxLifetime(time.Hour)

			return db
		}

		log.Printf("Failed to connect to database (attempt %d/30): %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("Could not connect to database after 30 attempts: %v", err)
	return nil
}

// migrateDatabase ejecuta las migraciones automáticas
func migrateDatabase(db *gorm.DB) {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migrations completed successfully")
}

// setupRouter configura las rutas de la API
func setupRouter(userController *controllers.UserController) *gin.Engine {
	router := gin.Default()

	// Middleware CORS (opcional pero útil para desarrollo frontend)
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "users-api",
		})
	})

	// Rutas públicas (sin autenticación)
	public := router.Group("/users")
	{
		public.POST("/register", userController.Register)
		public.POST("/login", userController.Login)
		public.POST("/admin", userController.RegisterAdmin) // Para crear el primer admin
	}

	// Rutas protegidas (requieren autenticación)
	protected := router.Group("/users")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/:id", userController.GetByID)
		protected.PUT("/:id", userController.Update)
	}

	// Rutas de administrador (requieren autenticación y rol admin)
	admin := router.Group("/users")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("", userController.GetAll)
		admin.DELETE("/:id", userController.Delete)
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
