package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"provider_management/internal/logger"
	"provider_management/internal/service"
)

type TransactionHandler struct {
	svc *service.TransactionService
	log *logger.Logger
}

func NewTransactionHandler(svc *service.TransactionService, log *logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		svc: svc,
		log: log,
	}
}

func (h *TransactionHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/transactions", h.GetAll)
	r.GET("/transactions/:id", h.GetByID)
}

func (h *TransactionHandler) GetAll(c *gin.Context) {
	res, err := h.svc.ListTransactions(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetTransaction(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}
