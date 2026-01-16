package handler

import (
	"net/http"
	"provider_management/internal/logger"
	"provider_management/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
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
	status := c.Query("status")
	method := c.Query("method")
	createdAt := c.Query("createdAt")
 

	res, total, err := h.svc.ListTransactions(c, page, limit, search, status, method, createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	pageInt, _ := strconv.ParseInt(page, 10, 64)
	limitInt, _ := strconv.ParseInt(limit, 10, 64)

	if pageInt < 1 {
		pageInt = 1
	}
	if limitInt < 1 {
		limitInt = 10
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + limitInt - 1) / limitInt
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
		"pagination": gin.H{
			"has_next":     pageInt < totalPages,
			"has_previous": pageInt > 1,
			"limit":        limitInt,
			"page":         pageInt,
			"total_items":  total,
			"total_pages":  totalPages,
		},
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
