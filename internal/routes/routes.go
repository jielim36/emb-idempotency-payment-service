package routes

import (
	"payment-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	paymentHandler *handlers.PaymentHandler,
	userHandler *handlers.UserHandler,
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"service": "emb-payment-service",
		})
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/pay", paymentHandler.ProcessPayment)
		v1.GET("/payments/transaction/:transactionId", paymentHandler.GetPaymentByTransactionID)
		v1.GET("/payments", paymentHandler.GetAll)

		userGrp := v1.Group("/users")
		{
			userGrp.GET("", userHandler.GetAll)
			userGrp.GET("/:userId", userHandler.GetDetail)
			userGrp.POST("/generate", userHandler.Generate)
		}
	}

	return router
}
