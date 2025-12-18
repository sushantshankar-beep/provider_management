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

func (h *TransactionHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	res, total, err := h.svc.ListTransactions(c, page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": total,
	})
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
