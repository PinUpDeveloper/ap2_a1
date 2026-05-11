package http

import (
	"errors"
	"net/http"
	"strconv"

	"ap2/order-service/internal/domain"
	"ap2/order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct{ usecase *usecase.OrderUsecase }

func NewOrderHandler(usecase *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: usecase}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required"`
}

func (h *OrderHandler) Register(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/recent", h.GetRecentOrders)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.usecase.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrPaymentUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	order, err := h.usecase.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	order, err := h.usecase.CancelOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrCancelNotAllowed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetRecentOrders(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	orders, err := h.usecase.GetRecentOrders(c.Request.Context(), limit)
	if err != nil {
		if err.Error() == "invalid limit" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orders == nil {
		orders = make([]domain.Order, 0)
	}
	c.JSON(http.StatusOK, orders)
}
