package handlers

import (
	"net/http"
	"payment-service/internal/models"
	"payment-service/internal/services"
	"payment-service/internal/utils/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	user, err := h.userService.GetAll()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all user", err)
		return
	}

	response.SuccessResponse(c, http.StatusOK, "success", user)
}

func (h *UserHandler) Generate(c *gin.Context) {
	user, wallet, err := h.userService.Generate()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all user", err)
		return
	}

	// temp use
	type details struct {
		User   *models.User
		Wallet *models.Wallet
	}

	response.SuccessResponse(c, http.StatusOK, "success", details{user, wallet})
}
