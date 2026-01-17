package handler

import (
	"log"
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

type ExportRequest struct {
	Format string   `json:"format"`
	IDs    []string `json:"settlement_ids"`
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

	settledBookingsResp := make([]gin.H, len(settlement.SettledBookings))
	for i, booking := range settlement.SettledBookings {
		settledBookingsResp[i] = gin.H{
			"service_id":         booking.ServiceID.Hex(),
			"service_request_no": booking.ServiceRequestNo,
			"original_amount":    booking.OriginalAmount,
			"settled_amount":     booking.SettledAmount,
			"settlement_type":    booking.SettlementType,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Settlement completed successfully",
		"data": gin.H{
			"settlement_id":    settlement.SettlementID,
			"payout_id":        settlement.PayoutID.Hex(),
			"provider_id":      settlement.ProviderID.Hex(),
			"provider_name":    settlement.ProviderName,
			"account_no":       settlement.AccountNo,
			"ifsc_code":        settlement.IfscCode,
			"total_amount":     settlement.TotalAmount,
			"payment_mode":     settlement.PaymentMode,
			"payment_method":   settlement.PaymentMethod,
			"justification":    settlement.Justification,
			"status":           settlement.Status,
			"settled_bookings": settledBookingsResp,
			"settled_at":       settlement.SettledAt,
			"created_at":       settlement.CreatedAt,
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
		"error":   false,
		"message": "Settlements fetched successfully",
		"data":    settlements,
		"pagination": gin.H{
			"total":        total,
			"total_pages":  totalPages,
			"current_page": req.Page,
			"limit":        req.Limit,
			"has_next":     req.Page < totalPages,
			"has_prev":     req.Page > 1,
		},
	})
}

func (h *SettlementHandler) ChangeProviderSettlementStatus(c *gin.Context) {
	settlementID := c.Param("id")
	if settlementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "settlement_id is required",
		})
		return
	}

	var req service.CreateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	_, err := h.svc.ChangeProviderSettlementStatus(c.Request.Context(), settlementID, &req)
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
	})
}

func (h *SettlementHandler) GetSettlementByID(c *gin.Context) {
	settlementID := c.Param("id")
	if settlementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "settlement_id is required",
		})
		return
	}

	settlement, err := h.svc.GetSettlementByID(c.Request.Context(), settlementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Settlement fetched successfully",
		"data":    settlement,
	})
}

func (h *SettlementHandler) ExportSettlements(c *gin.Context) {
	var req ExportRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"message": "No settlements selected"})
		return
	}

	format := req.Format
	if format == "" {
		format = "csv"
	}

	settlements, err := h.svc.GetSettlementsByIDs(
		c.Request.Context(),
		req.IDs,
	)
	log.Println("settlementssss", settlements)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	switch format {
	case "excel":
		h.exportExcel(c, settlements)
	case "pdf":
		h.exportPDF(c, settlements)
	default:
		h.exportCSV(c, settlements)
	}
}
