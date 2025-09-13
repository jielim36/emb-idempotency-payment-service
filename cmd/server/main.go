package main

import (
	"log"

	"payment-service/internal/config"
	"payment-service/internal/database"
	"payment-service/internal/handlers"
	"payment-service/internal/repositories"
	"payment-service/internal/routes"
	"payment-service/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	paymentRepo := repositories.NewPaymentRepository()

	// Initialize services
	paymentService := services.NewPaymentService(paymentRepo)

	// Initialize controllers
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// Setup routes
	router := routes.RegisterRoutes(paymentHandler)

	// Start server
	log.Printf("Starting %s v%s on port %s", cfg.App.Name, cfg.App.Version, cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
