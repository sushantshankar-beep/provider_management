package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"provider_management/internal/handler"
	"provider_management/internal/middleware"
	"time"
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
	serviceHandler *handler.ServiceHandler,
	adminHandler *handler.AdminHandler,
	authMiddleware *middleware.AuthMiddleware,
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
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	admin := r.Group("/admin")
	{
		users := admin.Group("/users")
		{
			users.GET("", userAdminHandler.GetAllUsers)
			users.GET("/:id", userAdminHandler.GetByID)
			users.PATCH("/:id/status", userAdminHandler.UpdateStatus)
		}

		providers := admin.Group("/providers")
		{
			providers.GET("", providerAdminHandler.GetAll)
			providers.GET("/:id", providerAdminHandler.GetByID)
			providers.PATCH("/status/:id", providerAdminHandler.UpdateStatus)
			providers.PATCH("/kyc/:id", providerAdminHandler.UpdateKYC)
			providers.PATCH("/verify-document/:id", providerAdminHandler.VerifyDocument)
			providers.PATCH("/account-action/:id", providerAdminHandler.UpdateAccountAction)
			providers.PATCH("/commission/:id", providerAdminHandler.UpdateCommission)
		}

		bookings := admin.Group("/bookings")
		{
			bookings.GET("", bookingAdminHandler.GetAllBookings)
			bookings.GET("/stats", bookingAdminHandler.GetBookingStats)
			bookings.GET("/:bookingId", bookingAdminHandler.GetBookingByID)
			bookings.GET("/get-invoice/:serviceId", bookingAdminHandler.GetInvoiceData)
			bookings.PUT("/:bookingId/cancel", bookingAdminHandler.CancelBooking)
			bookings.PUT("/:bookingId/complete", bookingAdminHandler.MarkBookingCompleted)
			bookings.POST("/:bookingId/notes", bookingAdminHandler.AddNote)
		}

		complaints := admin.Group("/complaints")
		{
			complaints.GET("", complaintHandler.GetAll)
			complaints.GET("/stats", complaintHandler.GetStats)
			complaints.GET("/:id", complaintHandler.GetByID)
			complaints.POST("/:id/assessment", complaintHandler.PostAssessment)
			complaints.POST("/:id/notes", complaintHandler.AddNote)
			complaints.PATCH("/:id/status", complaintHandler.UpdateStatus)
		}

		userTransactions := admin.Group("/user-transaction")
		{
			userTransactions.GET("", transactionHandler.GetAll)
			userTransactions.GET("/:id", transactionHandler.GetByID)
		}

		providerPayout := admin.Group("/provider-payout")
		{
			providerPayout.GET("", payoutHandler.GetPayouts)
			providerPayout.GET("/:id/details", payoutHandler.GetProviderPayoutDetails)
			providerPayout.GET("/:id/bookings", payoutHandler.GetPayoutServices)
			providerPayout.POST("/6hour", payoutHandler.Create6HourPayout)
		}

		providerSettlement := admin.Group("/provider-settlement")
		{
			providerSettlement.GET("", settlementHandler.GetSettlements)
			providerSettlement.POST("/create", settlementHandler.CreateSettlement)
		}

		services := admin.Group("/services")
		{
			services.POST("/create", serviceHandler.CreateService)
			services.GET("", serviceHandler.GetServices)
			services.GET("/stats", serviceHandler.GetServiceStats)
			services.GET("/:id", serviceHandler.GetServiceByID)
			services.PUT("/:id", serviceHandler.UpdateService)
			services.PATCH("/:id/status", serviceHandler.UpdateServiceStatus)
			services.DELETE("/:id", serviceHandler.DeleteService)
		}

		admin := admin.Group("/panel")
		{
			admin.POST("/login", adminHandler.Login)
			admin.POST("/logout", authMiddleware.AdminAuth(), adminHandler.Logout)
			admin.POST("/logout-all", authMiddleware.AdminAuth(), adminHandler.LogoutAll)
			admin.GET("/profile", authMiddleware.AdminAuth(), adminHandler.GetProfile)
			admin.POST("/change-password", authMiddleware.AdminAuth(), adminHandler.ChangeOwnPassword)
			
			admin.POST("/create", authMiddleware.AdminAuth(), adminHandler.CreateAdmin)
			admin.GET("/all", authMiddleware.AdminAuth(), adminHandler.GetAllAdmins)
			admin.GET("/stats", authMiddleware.AdminAuth(), adminHandler.GetDashboardStats)
			admin.GET("/:id", authMiddleware.AdminAuth(), adminHandler.GetAdminByID)
			admin.PUT("/:id", authMiddleware.AdminAuth(), adminHandler.UpdateAdmin)
			admin.PATCH("/:id/toggle-status", authMiddleware.AdminAuth(), adminHandler.ToggleAdminStatus)
			admin.DELETE("/:id", authMiddleware.AdminAuth(), adminHandler.DeleteAdmin)
			admin.POST("/:id/reset-password", authMiddleware.AdminAuth(), adminHandler.ResetPasswordBySuperAdmin)
		}
	}
}
