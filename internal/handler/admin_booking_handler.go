package handler

import (
	"log"
	"net/http"
	"provider_management/internal/dto"
	"provider_management/internal/middleware"
	"provider_management/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminBookingHandler struct {
	svc *service.AdminBookingService
}

func NewAdminBookingHandler(svc *service.AdminBookingService) *AdminBookingHandler {
	return &AdminBookingHandler{svc: svc}
}

func (h *AdminBookingHandler) GetAllBookings(c *gin.Context) {

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	filters := dto.BookingFilters{
		Search:        c.Query("search"),
		Status:        c.Query("status"),
		PaymentStatus: c.Query("paymentStatus"),
		ServiceType:   c.Query("serviceType"),
		VehicleType:   c.Query("vehicleType"),
		UserID:        c.Query("userId"),
		ProviderID:    c.Query("providerId"),
		BookingID:     c.Query("bookingId"),
		StartDate:     c.Query("startDate"),
		EndDate:       c.Query("endDate"),
		Sort:          c.Query("sort"),
	}

	zoneFilter := middleware.GetZoneFilter(c)

	pagination := dto.BookingPagination{
		Page:  page,
		Limit: limit,
		Sort:  c.DefaultQuery("sort", ""),
	}

	res, err := h.svc.GetAllBookings(c.Request.Context(), filters,pagination, zoneFilter)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch bookings",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Bookings fetched successfully",
		"data":    res,
	})
}

func (h *AdminBookingHandler) GetBookingStats(c *gin.Context) {
	params := make(map[string]string)
	for key, value := range c.Request.URL.Query() {
		if len(value) > 0 {
			params[key] = value[0]
		}
	}

	stats, err := h.svc.GetBookingStats(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch booking stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Booking stats fetched successfully",
		"data":    stats,
	})
}

func (h *AdminBookingHandler) GetBookingByID(c *gin.Context) {
	bookingID := c.Param("bookingId")

	res, err := h.svc.GetBookingByID(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Booking not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Booking fetched successfully",
		"data":    res,
	})
}

func (h *AdminBookingHandler) CancelBooking(c *gin.Context) {
	bookingID := c.Param("bookingId")

	res, err := h.svc.CancelBooking(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Booking cancelled successfully",
		"data":    res,
	})
}

func (h *AdminBookingHandler) MarkBookingCompleted(c *gin.Context) {
	bookingID := c.Param("bookingId")

	res, err := h.svc.MarkBookingCompleted(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Booking marked as completed successfully",
		"data":    res,
	})
}

func (h *AdminBookingHandler) AddNote(c *gin.Context) {
	bookingID := c.Param("bookingId")

	var req dto.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Note content is required",
		})
		return
	}

	if req.AddedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AddedBy field is required",
		})
		return
	}

	if err := h.svc.AddNote(c.Request.Context(), bookingID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add note: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Note added successfully",
	})
}

func (h *AdminBookingHandler) GetInvoice(c *gin.Context) {
	bookingID := c.Query("bookingId")
    log.Println("dlkcsnjsndjkdc",bookingID)
	invoice, err := h.svc.GetInvoice(c.Request.Context(), bookingID)
	log.Println("invoiceeee",invoice)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "invoice not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoice": invoice,
	})
}