package http

import (
	"errors"
	"net/http"

	"ap2/payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct{ usecase *usecase.PaymentUsecase }

func NewPaymentHandler(usecase *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{usecase: usecase}
}

type createPaymentRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	CustomerEmail string `json:"customer_email"`
}

func (h *PaymentHandler) Register(r *gin.Engine) {
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:order_id", h.GetPaymentByOrderID)
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, err := h.usecase.CreatePayment(c.Request.Context(), usecase.CreatePaymentInput{OrderID: req.OrderID, Amount: req.Amount, CustomerEmail: req.CustomerEmail})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidAmount), errors.Is(err, usecase.ErrOrderIDEmpty):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	payment, err := h.usecase.GetByOrderID(c.Request.Context(), c.Param("order_id"))
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}
