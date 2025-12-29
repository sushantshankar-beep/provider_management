package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"provider_management/internal/handler"
	"provider_management/internal/middleware"
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
	activityLogHandler *handler.ActivityLogHandler,
	amcPlanHandler *handler.AMCPlanHandler,
	amcTransactionHandler *handler.AMCTransactionHandler,
	amcOrderHandler *handler.OrderHandler,
	refundHandler *handler.RefundHandler,
	amcRefundHandler *handler.AMCRefundHandler,
	authMiddleware *middleware.AuthMiddleware,
	activityLogMiddleware *middleware.ActivityLogMiddleware,
	rbac *middleware.RBACMiddleware,
	roleHandler *handler.RoleHandler,
	zoneHandler *handler.ZoneHandler,
	vehicleBrandHandler *handler.VehicleBrandHandler,
	s3Uploader *middleware.S3Uploader,
) {

	// ===============================
	// CORS
	// ===============================
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

	// ===============================
	// ADMIN PANEL (LOGIN / PROFILE)
	// ===============================
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
			authenticated.POST("/create",
			s3Uploader.UploadMiddleware([]middleware.FieldConfig{
				{FormFieldName: "profileImage", ContextKey: "profileUrl"},
			}),
			adminHandler.CreateAdmin,
		)
			authenticated.GET("/all", adminHandler.GetAllAdmins)
			authenticated.GET("/stats", adminHandler.GetDashboardStats)
			authenticated.GET("/:id", adminHandler.GetAdminByID)
			authenticated.PUT("/:id",
				s3Uploader.UploadMiddleware([]middleware.FieldConfig{
					{FormFieldName: "profileImage", ContextKey: "profileUrl"},
				}),
				adminHandler.UpdateAdmin,
			)
			authenticated.PATCH("/:id/toggle-status", adminHandler.ToggleAdminStatus)
			authenticated.DELETE("/:id", adminHandler.DeleteAdmin)
			authenticated.POST("/:id/reset-password", adminHandler.ResetPasswordBySuperAdmin)
			authenticated.GET("/:id/activity-logs", activityLogHandler.GetAdminActivityLogs)
		}
	}


	// ===============================
	// ADMIN APIs
	// ===============================
	admin := r.Group("/admin")
	admin.Use(authMiddleware.AdminAuth())
	admin.Use(activityLogMiddleware.LogActivity())
	roles := admin.Group("/roles")
	{
		roles.POST(
			"",
			rbac.Check("rbac", "create"),
			roleHandler.CreateRole,
		)

		roles.GET(
			"",
			rbac.Check("rbac", "view"),
			roleHandler.ListRoles,
		)

		roles.PATCH(
			"/:id/status",
			rbac.Check("rbac", "update"),
			roleHandler.ToggleRoleStatus,
		)
		roles.POST(
			"/:id/clone",
			rbac.Check("rbac", "create"),
			roleHandler.CloneRole,
		)
		roles.DELETE(
			"/:id",
			rbac.Check("rbac", "delete"),
			roleHandler.DeleteRole,
		)
		roles.PUT(
			"/:id",
			rbac.Check("rbac", "update"),
			roleHandler.UpdateRole,
		)
		roles.GET(
			"/:id",
			rbac.Check("rbac", "view"),
			roleHandler.GetRoleByID,
		)
	}

	// ---------- USERS ----------
	users := admin.Group("/users")
	{
		users.GET("", rbac.Check("users", "view"), userAdminHandler.GetAllUsers)
		users.GET("/:id", rbac.Check("users", "view"), userAdminHandler.GetByID)
		users.GET("/:id/activity-logs", rbac.Check("activity_logs", "view"), activityLogHandler.GetUserActivityLogs)
		users.PATCH("/:id/status", rbac.Check("users", "status"), userAdminHandler.UpdateStatus)
	}

	// ---------- PROVIDERS ----------
	providers := admin.Group("/providers")
	{
		providers.GET("", rbac.Check("providers", "view"), providerAdminHandler.GetAll)
		providers.GET("/:id", rbac.Check("providers", "view"), providerAdminHandler.GetByID)
		providers.GET("/:id/activity-logs", rbac.Check("activity_logs", "view"), activityLogHandler.GetProviderActivityLogs)

		providers.PATCH("/status/:id", rbac.Check("providers", "status"), providerAdminHandler.UpdateStatus)
		providers.PATCH("/kyc/:id", rbac.Check("providers", "kyc"), providerAdminHandler.UpdateKYC)
		providers.PATCH("/verify-document/:id", rbac.Check("providers", "verify-document"), providerAdminHandler.VerifyDocument)
		providers.PATCH("/account-action/:id", rbac.Check("providers", "account-action"), providerAdminHandler.UpdateAccountAction)
		providers.PATCH("/commission/:id", rbac.Check("providers", "commission"), providerAdminHandler.UpdateCommission)
	}

	// ---------- BOOKINGS ----------
	bookings := admin.Group("/bookings")
	{
		bookings.GET("", rbac.Check("bookings", "view"), bookingAdminHandler.GetAllBookings)
		bookings.GET("/stats", rbac.Check("bookings", "stats"), bookingAdminHandler.GetBookingStats)
		bookings.GET("/:bookingId", rbac.Check("bookings", "view"), bookingAdminHandler.GetBookingByID)
		bookings.GET("/:bookingId/activity-logs", rbac.Check("activity_logs", "view"), activityLogHandler.GetBookingActivityLogs)
		bookings.GET("/get-invoice/:serviceId", rbac.Check("bookings", "invoice"), bookingAdminHandler.GetInvoiceData)
		providerSettlement := admin.Group("/provider-settlement")
		{
			providerSettlement.GET("", settlementHandler.GetSettlements)
			providerSettlement.POST("/create", settlementHandler.CreateSettlement)
			providerSettlement.POST("/:id/settle", settlementHandler.ChangeProviderSettlementStatus)
			providerSettlement.GET("/:id", settlementHandler.GetSettlementByID)
			providerSettlement.GET("/export", settlementHandler.ExportSettlements)
		}

		bookings.PUT("/:bookingId/cancel", rbac.Check("bookings", "cancel"), bookingAdminHandler.CancelBooking)
		bookings.PUT("/:bookingId/complete", rbac.Check("bookings", "complete"), bookingAdminHandler.MarkBookingCompleted)
		bookings.POST("/:bookingId/notes", rbac.Check("bookings", "notes"), bookingAdminHandler.AddNote)
	}

	// ---------- COMPLAINTS ----------
	complaints := admin.Group("/complaints")
	{
		complaints.GET("", rbac.Check("complaints", "view"), complaintHandler.GetAll)
		complaints.GET("/stats", rbac.Check("complaints", "stats"), complaintHandler.GetStats)
		complaints.GET("/:id", rbac.Check("complaints", "view"), complaintHandler.GetByID)
		complaints.GET("/:id/activity-logs", rbac.Check("activity_logs", "view"), activityLogHandler.GetComplaintActivityLogs)

		complaints.POST("/:id/assessment", rbac.Check("complaints", "assessment"), complaintHandler.PostAssessment)
		complaints.POST("/:id/notes", rbac.Check("complaints", "notes"), complaintHandler.AddNote)
		complaints.PATCH("/:id/status", rbac.Check("complaints", "status"), complaintHandler.UpdateStatus)
	}

	// ---------- TRANSACTIONS ----------
	userTransactions := admin.Group("/user-transaction")
	{
		userTransactions.GET("", rbac.Check("transactions", "view"), transactionHandler.GetAll)
		userTransactions.GET("/:id", rbac.Check("transactions", "view"), transactionHandler.GetByID)
	}

	// ---------- PAYOUTS ----------
	providerPayout := admin.Group("/provider-payout")
	{
		providerPayout.GET("", rbac.Check("payouts", "view"), payoutHandler.GetPayouts)
		providerPayout.GET("/:id/details", rbac.Check("payouts", "view"), payoutHandler.GetProviderPayoutDetails)
		providerPayout.GET("/:id/bookings", rbac.Check("payouts", "view"), payoutHandler.GetPayoutServices)
		providerPayout.POST("/6hour", rbac.Check("payouts", "create"), payoutHandler.Create6HourPayout)
	}

	// ---------- SETTLEMENT ----------
	providerSettlement := admin.Group("/provider-settlement")
	{
		providerSettlement.GET("", rbac.Check("settlement", "view"), settlementHandler.GetSettlements)
		providerSettlement.POST("/create", rbac.Check("settlement", "create"), settlementHandler.CreateSettlement)
		providerSettlement.POST("/:id/settle", rbac.Check("settlement", "update"), settlementHandler.ChangeProviderSettlementStatus)
		providerSettlement.GET("/:id", rbac.Check("settlement", "view"), settlementHandler.GetSettlementByID)
		providerSettlement.POST("/export", rbac.Check("settlement", "export"), settlementHandler.ExportSettlements)
	}
	// ---------- SERVICES ----------
	services := admin.Group("/services")
	{
		services.POST("/create", rbac.Check("services", "create"), serviceHandler.CreateService)
		services.GET("", rbac.Check("services", "view"), serviceHandler.GetServices)
		services.GET("/stats", rbac.Check("services", "stats"), serviceHandler.GetServiceStats)
		services.GET("/:id", rbac.Check("services", "view"), serviceHandler.GetServiceByID)
		services.GET("/:id/activity-logs", rbac.Check("activity_logs", "view"), activityLogHandler.GetServiceActivityLogs)
		services.PUT("/:id", rbac.Check("services", "update"), serviceHandler.UpdateService)
		services.PATCH("/:id/status", rbac.Check("services", "status"), serviceHandler.UpdateServiceStatus)
		services.DELETE("/:id", rbac.Check("services", "delete"), serviceHandler.DeleteService)
	}

	// ---------- ACTIVITY LOGS ----------
	activityLogs := admin.Group("/activity-logs")
	{
		activityLogs.GET("", rbac.Check("activity_logs", "view"), activityLogHandler.GetAllActivityLogs)
	}

	// ---------- AMC ----------
	amcPlans := admin.Group("/amc-plans")
	{
		amcPlans.POST("/create", rbac.Check("amc", "create"), amcPlanHandler.CreateAMC)
		amcPlans.GET("", rbac.Check("amc", "view"), amcPlanHandler.GetAllAMC)
		amcPlans.GET("/:id", rbac.Check("amc", "view"), amcPlanHandler.GetAMCByID)
		amcPlans.PUT("/:id", rbac.Check("amc", "update"), amcPlanHandler.UpdateAMC)
		amcPlans.DELETE("/:id", rbac.Check("amc", "delete"), amcPlanHandler.DeleteAMC)
		amcPlans.PATCH("/:id/toggle-status", rbac.Check("amc", "status"), amcPlanHandler.ToggleAMCStatus)
	}

	amcTransaction := admin.Group("/amc-transaction")
	{
		amcTransaction.GET("", rbac.Check("amc", "view"), amcTransactionHandler.GetAll)
		amcTransaction.GET("/:id", rbac.Check("amc", "view"), amcTransactionHandler.GetByID)
	}

	amcOrder := admin.Group("/amc-order")
	{
		amcOrder.GET("", rbac.Check("amc", "view"), amcOrderHandler.GetAll)
		amcOrder.GET("/export", rbac.Check("amc", "export"), amcOrderHandler.ExportToCSV)
		amcOrder.GET("/:id", rbac.Check("amc", "view"), amcOrderHandler.GetByID)
		amcOrder.PATCH("/:id/status", rbac.Check("amc", "status"), amcOrderHandler.UpdateStatus)
	}

	userRefund := admin.Group("/user-refund")
	{
		userRefund.GET("", rbac.Check("refunds", "view"), refundHandler.GetAllRefunds)
		userRefund.GET("/:id", rbac.Check("refunds", "view"), refundHandler.GetRefundByID)
	}

	amcRefund := admin.Group("/amc-refund")
	{
		amcRefund.GET("", rbac.Check("refunds", "view"), amcRefundHandler.GetAll)
		amcRefund.GET("/:id", rbac.Check("refunds", "view"), amcRefundHandler.GetDetails)
		amcRefund.PUT("/:id/approve", rbac.Check("refunds", "approve"), amcRefundHandler.Approve)
		amcRefund.PUT("/:id/reject", rbac.Check("refunds", "reject"), amcRefundHandler.Reject)
		amcRefund.GET("/:id/check-status", rbac.Check("refunds", "view"), amcRefundHandler.CheckStatus)
		}

		zones := admin.Group("/zones")
		{
			zones.POST("", zoneHandler.Create)
			zones.GET("", zoneHandler.GetAll)
			zones.GET("/active", zoneHandler.GetActive)
			zones.PUT("/:id", zoneHandler.Update)
			zones.PATCH("/:id/toggle-status", zoneHandler.ToggleStatus)
			zones.DELETE("/:id", zoneHandler.Delete)
		}

		vehicleBrands := admin.Group("/vehicle-brands")
		{
			vehicleBrands.GET("", vehicleBrandHandler.GetBrands)
			vehicleBrands.GET("/models", vehicleBrandHandler.GetModels)
		}
	}
}
