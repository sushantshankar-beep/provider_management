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
	activityLogMiddleware *middleware.ActivityLogMiddleware,
	activityLogHandler *handler.ActivityLogHandler,
	amcPlanHandler *handler.AMCPlanHandler,
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

	adminPanel := r.Group("/admin/panel")
	{
		adminPanel.POST("/login", adminHandler.Login)
		
		authenticated := adminPanel.Group("")
		authenticated.Use(authMiddleware.AdminAuth())
		authenticated.Use(activityLogMiddleware.LogActivity())
		{
			authenticated.POST("/logout", adminHandler.Logout)
			authenticated.POST("/logout-all", adminHandler.LogoutAll)
			authenticated.GET("/profile", adminHandler.GetProfile)
			authenticated.POST("/change-password", adminHandler.ChangeOwnPassword)
			authenticated.POST("/create", adminHandler.CreateAdmin)
			authenticated.GET("/all", adminHandler.GetAllAdmins)
			authenticated.GET("/stats", adminHandler.GetDashboardStats)
			authenticated.GET("/:id", adminHandler.GetAdminByID)
			authenticated.PUT("/:id", adminHandler.UpdateAdmin)
			authenticated.PATCH("/:id/toggle-status", adminHandler.ToggleAdminStatus)
			authenticated.DELETE("/:id", adminHandler.DeleteAdmin)
			authenticated.POST("/:id/reset-password", adminHandler.ResetPasswordBySuperAdmin)
		}
	}

	admin := r.Group("/admin")
	admin.Use(authMiddleware.AdminAuth())
	admin.Use(activityLogMiddleware.LogActivity())
	{
		users := admin.Group("/users")
		{
			users.GET("", userAdminHandler.GetAllUsers)
			users.GET("/:id", userAdminHandler.GetByID)
			users.GET("/:id/activity-logs", activityLogHandler.GetUserActivityLogs)
			users.PATCH("/:id/status", userAdminHandler.UpdateStatus)
		}

		providers := admin.Group("/providers")
		{
			providers.GET("", providerAdminHandler.GetAll)
			providers.GET("/:id", providerAdminHandler.GetByID)
			providers.GET("/:id/activity-logs", activityLogHandler.GetProviderActivityLogs)
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
			bookings.GET("/:bookingId/activity-logs", activityLogHandler.GetBookingActivityLogs)
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
			complaints.GET("/:id/activity-logs", activityLogHandler.GetComplaintActivityLogs)
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
			services.GET("/:id/activity-logs", activityLogHandler.GetServiceActivityLogs)
			services.PUT("/:id", serviceHandler.UpdateService)
			services.PATCH("/:id/status", serviceHandler.UpdateServiceStatus)
			services.DELETE("/:id", serviceHandler.DeleteService)
		}



		activityLogs := admin.Group("/activity-logs")
		{
			activityLogs.GET("", activityLogHandler.GetAllActivityLogs)
		}
		
		amcPlans := admin.Group("/amc-plans")
		{
			amcPlans.POST("/create", amcPlanHandler.CreateAMC)
			amcPlans.GET("", amcPlanHandler.GetAllAMC)
			amcPlans.GET("/:id", amcPlanHandler.GetAMCByID)
			amcPlans.PUT("/:id", amcPlanHandler.UpdateAMC)
			amcPlans.DELETE("/:id", amcPlanHandler.DeleteAMC)
			amcPlans.PATCH("/:id/toggle-status", amcPlanHandler.ToggleAMCStatus)
		}
	}
}