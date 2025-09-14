package main

import (
	"log"

	"payment-service/internal/config"
	"payment-service/internal/database"
	"payment-service/internal/handlers"
	"payment-service/internal/repositories"
	"payment-service/internal/routes"
	"payment-service/internal/services"
	"payment-service/internal/validator"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// TODO: Implement DI Container for future
	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	db := database.GetDB()
	paymentRepo := repositories.NewPaymentRepository(db)
	walletRepo := repositories.NewWalletRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Initialize services
	paymentService := services.NewPaymentService(db, paymentRepo, walletRepo)
	userService := services.NewUserService(db, userRepo, walletRepo)

	// Initialize controllers
	paymentHandler := handlers.NewPaymentHandler(paymentService, userService)
	userHandler := handlers.NewUserHandler(userService)

	// Setup routes
	router := routes.RegisterRoutes(paymentHandler, userHandler)

	// Register validators
	validator.RegisterValidators()

	// Start server
	log.Printf("Starting %s v%s on port %s", cfg.App.Name, cfg.App.Version, cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
