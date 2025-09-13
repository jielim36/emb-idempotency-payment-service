package routes

import (
	"payment-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(paymentHandler *handlers.PaymentHandler) *gin.Engine {
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
	}

	return router
}
