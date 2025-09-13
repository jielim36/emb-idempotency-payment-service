package handlers

import (
	"errors"
	"net/http"

	"payment-service/internal/models"
	"payment-service/internal/services"
	"payment-service/internal/utils/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req models.PaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err)
		return
	}

	payment, err := h.paymentService.ProcessPayment(c, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Failed to process payment", err)
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Payment initiated successfully", payment)
}

func (h *PaymentHandler) GetPaymentByTransactionID(c *gin.Context) {
	txId := c.Param("transactionId")
	payment, err := h.paymentService.GetPaymentByTransactionID(txId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "Failed to get payment by transaction id", err)
		} else {
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payment by transaction id", err)
		}
		return
	}

	response.SuccessResponse(c, http.StatusOK, "success", payment)
}

func (h *PaymentHandler) GetAll(c *gin.Context) {
	payment, err := h.paymentService.GetAll()
	// if err != nil {
	// 	response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all payment", err)
	// 	return
	// }

	response.SuccessResponse(c, http.StatusOK, "success", payment)
}
