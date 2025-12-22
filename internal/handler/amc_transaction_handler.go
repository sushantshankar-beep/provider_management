package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"provider_management/internal/service"
	"strconv"
)

type AMCTransactionHandler struct {
	svc *service.AMCTransactionService
}

func NewAMCTransactionHandler(svc *service.AMCTransactionService) *AMCTransactionHandler {
	return &AMCTransactionHandler{
		svc: svc,
	}
}

func (h *AMCTransactionHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	status := c.Query("status")
	method := c.Query("method")

	res, total, err := h.svc.ListAMCTransactions(c, page, limit, search, status, method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error fetching transactions", "error": err.Error()})
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
		"success": true,
		"data":    res,
		"pagination": gin.H{
			"total":      total,
			"page":       pageInt,
			"limit":      limitInt,
			"totalPages": totalPages,
		},
	})
}

func (h *AMCTransactionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetAMCTransaction(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Transaction not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}
