package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"provider_management/internal/service"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		svc: svc,
	}
}

func (h *TransactionHandler) GetAll(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	filters := dto.TransactionFilters{
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		Method:    c.Query("method"),
		CreatedAt: c.Query("createdAt"),
	}
 
	pagination := dto.PaginationParams{
		Page:  page,
		Limit: limit,
	}

	res, err := h.svc.GetTransactions(c, filters, pagination)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, res)

}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetTransactionById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}
