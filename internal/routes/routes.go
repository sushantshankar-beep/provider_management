package routes

import (
	"time"

	"provider_management/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	allowedOrigins []string,
	complaintHandler *handler.ComplaintHandler,
	transactionHandler *handler.TransactionHandler,
	userAdminHandler *handler.UserAdminHandler,
	providerAdminHandler *handler.ProviderAdminHandler,
	bookingAdminHandler *handler.AdminBookingHandler,
	payoutHandler *handler.PayoutHandler,
    settlementHandler *handler.SettlementHandler,
) {
	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"ngrok-skip-browser-warning",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
    
	r.GET("/complaints", complaintHandler.GetAll)
	r.GET("/complaints/:id", complaintHandler.GetByID)
	r.POST("/complaints/:id/assessment", complaintHandler.PostAssessment)

	r.GET("/transactions", transactionHandler.GetAll)
	r.GET("/transactions/:id", transactionHandler.GetByID)

	r.GET("/admin/users", userAdminHandler.GetAll)
	r.GET("/admin/users/:id", userAdminHandler.GetByID)
	r.PATCH("/admin/users/:id/status", userAdminHandler.UpdateStatus)

	r.GET("/admin/providers", providerAdminHandler.GetAll)
	r.GET("/admin/providers/:id", providerAdminHandler.GetByID)
	r.PATCH("/admin/providers/status/:id", providerAdminHandler.UpdateStatus)
	r.PATCH("/admin/providers/kyc/:id", providerAdminHandler.UpdateKYC)
	r.PATCH("/admin/providers/verify-document/:id", providerAdminHandler.VerifyDocument)
	r.PATCH("/admin/providers/:id/account-action", providerAdminHandler.UpdateAccountAction)
	r.PATCH("/admin/providers/commission/:id", providerAdminHandler.UpdateCommission)

	r.GET("/admin/bookings", bookingAdminHandler.GetAllBookings)
    r.GET("/admin/bookings/stats", bookingAdminHandler.GetBookingStats)
    r.GET("/admin/bookings/:bookingId", bookingAdminHandler.GetBookingByID)
    r.PUT("/admin/bookings/:bookingId/cancel",bookingAdminHandler.CancelBooking)
    r.PUT("/admin/bookings/:bookingId/complete", bookingAdminHandler.MarkBookingCompleted)
    r.GET("/admin/bookings/get-invoice/:serviceId", bookingAdminHandler.GetInvoiceData)

	r.POST("/admin/payouts/6hour",payoutHandler.Create6HourPayout)
    r.GET("/admin/payouts", payoutHandler.GetPayouts)
    r.GET("/admin/payouts/:id/services", payoutHandler.GetPayoutServices)
    r.GET("/admin/payouts/:id/provider", payoutHandler.GetPayoutProviderData)

    r.POST("/settlements", settlementHandler.CreateSettlement)
    r.GET("/allsettlements", settlementHandler.GetSettlements)

}
