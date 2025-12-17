package handler

import (
	"net/http"
	"provider_management/internal/service"
	"github.com/gin-gonic/gin"

)

type SettlementHandler struct {
	svc *service.SettlementService
}

func NewSettlementHandler(svc *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{svc: svc}
}



func (h *SettlementHandler) CreateSettlement(c *gin.Context) {
	var req service.SettlementRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	settlement, err := h.svc.CreateSettlement(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Settlement completed successfully",
		"data": gin.H{
			"settlement_id": settlement.SettlementID,
			"payout_id":     settlement.PayoutID.Hex(),
			"provider_id":   settlement.ProviderID.Hex(),
			"provider_name": settlement.ProviderName,
			"account_no":    settlement.AccountNo,
			"ifsc_code":     settlement.IfscCode,
			"total_amount":  settlement.TotalAmount,
			"payment_mode":   settlement.PaymentMode,
			"payment_method": settlement.PaymentMethod,
			"justification": settlement.Justification,
			"status":        settlement.Status,
			"settled_at":    settlement.SettledAt,
			"created_at":    settlement.CreatedAt,

		},
	})
}

func (h *SettlementHandler) GetSettlements(c *gin.Context) {
	var req service.GetSettlementsRequest
	
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters: " + err.Error(),
		})
		return
	}

	settlements, total, totalPages, err := h.svc.GetSettlements(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":       false,
		"message":     "Settlements fetched successfully",
		"data":        settlements,
		"pagination": gin.H{
			"total":       total,
			"total_pages": totalPages,
			"current_page": req.Page,
			"limit":       req.Limit,
			"has_next":    req.Page < totalPages,
			"has_prev":    req.Page > 1,
		},
	})
}
