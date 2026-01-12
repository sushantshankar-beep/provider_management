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
	dashboardHandler *handler.DashboardHandler,
	s3Uploader *middleware.S3Uploader,
	zoneFilter *middleware.ZoneFilterMiddleware,
	permissionHandler *handler.PermissionHandler,
	zoneMapHandler *handler.ZoneMapHandler,
	vehicleHandler *handler.VehicleHandler,
) {
	r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

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
			authenticated.POST("/change-password", rbac.Check("admins", "change_password"), adminHandler.ChangeOwnPassword)
			authenticated.POST("/create",
				s3Uploader.UploadMiddleware([]middleware.FieldConfig{
					{FormFieldName: "profileImage", ContextKey: "profileUrl"},
				}),
				rbac.Check("admins", "create_admin/sub-admin"),
				adminHandler.CreateAdmin,
			)
			authenticated.GET("/all", rbac.Check("admins", "view_admin/sub-admin"), adminHandler.GetAllAdmins)
			authenticated.GET("/stats", rbac.Check("admins", "view_admin/sub-admin"), adminHandler.GetDashboardStats)
			authenticated.GET("/:id", rbac.Check("admins", "view_admin/sub-admin_details"), adminHandler.GetAdminByID)
			authenticated.PUT("/:id",
				s3Uploader.UploadMiddleware([]middleware.FieldConfig{
					{FormFieldName: "profileImage", ContextKey: "profileUrl"},
				}),
				rbac.Check("admins", "edit_admin/sub-admin"),
				adminHandler.UpdateAdmin,
			)
			authenticated.PATCH("/:id/toggle-status", rbac.Check("admins", "edit_admin/sub-admin"), adminHandler.ToggleAdminStatus)
			authenticated.DELETE("/:id", rbac.Check("admins", "delete_admin/subadmin"), adminHandler.DeleteAdmin)
			authenticated.POST("/:id/reset-password", adminHandler.ResetPasswordBySuperAdmin)
			authenticated.GET("/:id/activity-logs", rbac.Check("admins", "admin/sub-admin_activity_logs"), activityLogHandler.GetAdminActivityLogs)
		}
	}

	admin := r.Group("/admin")
	admin.Use(authMiddleware.AdminAuth())
	admin.Use(activityLogMiddleware.LogActivity())
	roles := admin.Group("/roles")
	{
		roles.GET("/my-role", roleHandler.GetMyRole)
		roles.POST(
			"",
			rbac.Check("roles", "create_new_role"),
			roleHandler.CreateRole,
		)

		roles.GET(
			"",
			rbac.Check("roles", "view_role"),
			roleHandler.ListRoles,
		)

		roles.PATCH(
			"/:id/status",
			rbac.Check("roles", "update_roles"),
			roleHandler.ToggleRoleStatus,
		)
		roles.POST(
			"/:id/clone",
			rbac.Check("roles", "create_new_role"),
			roleHandler.CloneRole,
		)
		roles.DELETE(
			"/:id",
			rbac.Check("roles", "delete_roles"),
			roleHandler.DeleteRole,
		)
		roles.PUT(
			"/:id",
			rbac.Check("roles", "update_roles"),
			roleHandler.UpdateRole,
		)
		roles.GET(
			"/:id",
			rbac.Check("roles", "view_role_details"),
			roleHandler.GetRoleByID,
		)
		roles.GET("/role-type", roleHandler.GetRoleTypes)
		roles.GET("/role-names", roleHandler.GetRoleNamesByType)
	}

	dashboard := admin.Group("/dashboard")
	{
		dashboard.GET("/stats", rbac.Check("dashboard", "view_dashboard"), dashboardHandler.GetDashboardStats)
	}

	users := admin.Group("/users")
	users.Use(zoneFilter.ApplyZoneFilter("zoneScope"))
	{
		users.GET("", rbac.Check("users", "view_users"), userAdminHandler.GetAllUsers)
		users.GET("/:id", rbac.Check("users", "view_user_profile"), userAdminHandler.GetByID)
		users.GET("/:id/activity-logs", rbac.Check("users", "view_user_activity_logs"), activityLogHandler.GetUserActivityLogs)
		users.PATCH("/:id/status", rbac.Check("users", "change_user_status"), userAdminHandler.UpdateStatus)
		users.POST("/:id/notes",rbac.Check("users", "add_user_notes"), userAdminHandler.AddNote)
	}

	providers := admin.Group("/providers")
	providers.Use(zoneFilter.ApplyZoneFilter("zoneName"))
	{
		providers.GET("", rbac.Check("providers", "view_providers"), providerAdminHandler.GetAll)
		providers.GET("/:id", rbac.Check("providers", "view_provider_profile"), providerAdminHandler.GetByID)
		providers.POST("",
		    rbac.Check("providers", "create_provider_profile"),
			s3Uploader.UploadMiddleware([]middleware.FieldConfig{
				{FormFieldName: "profileImage", ContextKey: "profileUrl"},
				{FormFieldName: "identityProof", ContextKey: "identityProof"},
				{FormFieldName: "addressProof", ContextKey: "addressProof"},
				{FormFieldName: "cancelCheque", ContextKey: "cancelCheque"},
			}),
			providerAdminHandler.CreateProvider,
		)

		providers.PUT("/:id",
		    rbac.Check("providers", "edit_provider_profile"),
			s3Uploader.UploadMiddleware([]middleware.FieldConfig{
				{FormFieldName: "profileImage", ContextKey: "profileUrl"},
				{FormFieldName: "identityProof", ContextKey: "identityProof"},
				{FormFieldName: "addressProof", ContextKey: "addressProof"},
				{FormFieldName: "cancelCheque", ContextKey: "cancelCheque"},
			}),
			providerAdminHandler.UpdateProvider,
		)

		providers.GET("/:id/activity-logs", rbac.Check("providers", "view_provider_activity_logs"), activityLogHandler.GetProviderActivityLogs)
		providers.PATCH("/status/:id", rbac.Check("providers", "change_provider_status"), providerAdminHandler.UpdateStatus)
		providers.PATCH("/kyc/:id", rbac.Check("providers", "kyc"), providerAdminHandler.UpdateKYC)
		providers.PATCH("/verify-document/:id", rbac.Check("providers", "verify_provider_document"), providerAdminHandler.VerifyDocument)
		providers.PATCH("/account-action/:id", rbac.Check("providers", "change_provider_status"), providerAdminHandler.UpdateAccountAction)
		providers.PATCH("/commission/:id", rbac.Check("providers", "change_provider_commission"), providerAdminHandler.UpdateCommission)
		providers.GET("/:id/documents/download", rbac.Check("providers", "view_provider_profile"), providerAdminHandler.DownloadDocument)
		providers.POST("/:id/notes",  rbac.Check("providers", "edit_provider_profile"),providerAdminHandler.AddNote)
		providers.GET("/zones/stats", rbac.Check("providers", "view_providers"), providerAdminHandler.GetZoneStats)
		providers.GET("/zones/:zone/activation-team", rbac.Check("providers", "view_providers"), providerAdminHandler.GetZoneActivationTeam)
		providers.GET("/zones/:zone/activation-team/:person", rbac.Check("providers", "view_providers"), providerAdminHandler.GetActivationPersonProviders)
		providers.GET("/activation-team/:person", rbac.Check("providers", "view_providers"), providerAdminHandler.GetActivationPersonProviders)

	}

	bookings := admin.Group("/bookings")
	bookings.Use(zoneFilter.ApplyZoneFilter("both"))
	{
		bookings.GET("", rbac.Check("bookings", "view_bookings"), bookingAdminHandler.GetAllBookings)
		bookings.GET("/stats", rbac.Check("bookings", "view_bookings"), bookingAdminHandler.GetBookingStats)
		bookings.GET("/:bookingId", rbac.Check("bookings", "view_booking_profile"), bookingAdminHandler.GetBookingByID)
		bookings.GET("/:bookingId/activity-logs", rbac.Check("bookings", "view_booking_activity_logs"), activityLogHandler.GetBookingActivityLogs)
		bookings.GET("/get-invoice/:serviceId", rbac.Check("bookings", "view_booking_profile"), bookingAdminHandler.GetInvoiceData)
		bookings.PUT("/:bookingId/cancel", rbac.Check("bookings", "booking_action"), bookingAdminHandler.CancelBooking)
		bookings.PUT("/:bookingId/complete", rbac.Check("bookings", "booking_action"), bookingAdminHandler.MarkBookingCompleted)
		bookings.POST("/:bookingId/notes", rbac.Check("bookings", "add_booking_notes"), bookingAdminHandler.AddNote)
	}

	complaints := admin.Group("/complaints")
	{
		complaints.GET("", rbac.Check("complaints", "view_complaint"), complaintHandler.GetAll)
		complaints.GET("/stats", rbac.Check("complaints", "view_complaint"), complaintHandler.GetStats)
		complaints.GET("/:id", rbac.Check("complaints", "view_complaint_details"), complaintHandler.GetByID)
		complaints.GET("/:id/activity-logs", rbac.Check("complaints", "view_complaint_activity_logs"), activityLogHandler.GetComplaintActivityLogs)
		complaints.POST("/:id/start-assessment", rbac.Check("complaints", "perform_assessment"), complaintHandler.StartAssessment)
		complaints.POST("/:id/assessment", rbac.Check("complaints", "perform_assessment"), complaintHandler.PostAssessment)
		complaints.POST("/:id/notes", rbac.Check("complaints", "add_complaint_notes"), complaintHandler.AddNote)
		complaints.PATCH("/:id/status", rbac.Check("complaints", "status"), complaintHandler.UpdateStatus)
	}

	userTransactions := admin.Group("/user-transaction")
	{
		userTransactions.GET("", rbac.Check("paymentandtransactions", "view_users_transaction"), transactionHandler.GetAll)
		userTransactions.GET("/:id", rbac.Check("paymentandtransactions", "view_user_transaction_details"), transactionHandler.GetByID)
	}

	providerPayout := admin.Group("/provider-payout")
	{
		providerPayout.GET("", rbac.Check("paymentandtransactions", "view_provider_payout"), payoutHandler.GetPayouts)
		providerPayout.GET("/:id/details", rbac.Check("paymentandtransactions", "view_provider_payout_details"), payoutHandler.GetProviderPayoutDetails)
		providerPayout.GET("/:id/bookings", rbac.Check("paymentandtransactions", "view_provider_payout_details"), payoutHandler.GetPayoutServices)
		providerPayout.POST("/6hour", rbac.Check("paymentandtransactions", "create"), payoutHandler.Create6HourPayout)
	}

	providerSettlement := admin.Group("/provider-settlement")
	{
		providerSettlement.GET("", rbac.Check("paymentandtransactions", "view_provider_settlement"), settlementHandler.GetSettlements)
		providerSettlement.POST("/create", rbac.Check("paymentandtransactions", "create_provider_settlement"), settlementHandler.CreateSettlement)
		providerSettlement.POST("/:id/settle", rbac.Check("paymentandtransactions", "process_final_settlement"), settlementHandler.ChangeProviderSettlementStatus)
		providerSettlement.GET("/:id", rbac.Check("paymentandtransactions", "check_final_settlement"), settlementHandler.GetSettlementByID)
		providerSettlement.POST("/export", rbac.Check("paymentandtransactions", "export"), settlementHandler.ExportSettlements)
	}

	services := admin.Group("/services")
	{
		services.POST("/create", rbac.Check("services", "create_service"), serviceHandler.CreateService)
		services.GET("", rbac.Check("services", "view_service"), serviceHandler.GetServices)
		services.GET("/stats", rbac.Check("services", "view_service"), serviceHandler.GetServiceStats)
		services.GET("/:id", rbac.Check("services", "view_service_details"), serviceHandler.GetServiceByID)
		services.GET("/:id/activity-logs", rbac.Check("services", "service_master_activity_logs"), activityLogHandler.GetServiceActivityLogs)
		services.PUT("/:id", rbac.Check("services", "edit_service"), serviceHandler.UpdateService)
		services.PATCH("/:id/status", rbac.Check("services", "edit_service"), serviceHandler.UpdateServiceStatus)
		services.DELETE("/:id", rbac.Check("services", "delete_service"), serviceHandler.DeleteService)
	}

	activityLogs := admin.Group("/activity-logs")
	{
		activityLogs.GET("", rbac.Check("activity_logs", "view"), activityLogHandler.GetAllActivityLogs)
	}

	amcPlans := admin.Group("/amc-plans")
	{
		amcPlans.POST("/create", rbac.Check("amc", "create_amc_plan"), amcPlanHandler.CreateAMC)
		amcPlans.GET("", rbac.Check("amc", "view_amc_plan"), amcPlanHandler.GetAllAMC)
		amcPlans.GET("/:id", rbac.Check("amc", "view_amc_detail"), amcPlanHandler.GetAMCByID)
		amcPlans.PUT("/:id", rbac.Check("amc", "edit_amc_plan"), amcPlanHandler.UpdateAMC)
		amcPlans.DELETE("/:id", rbac.Check("amc", "delete_amc_plan"), amcPlanHandler.DeleteAMC)
		amcPlans.PATCH("/:id/toggle-status", rbac.Check("amc", "edit_amc_plan"), amcPlanHandler.ToggleAMCStatus)
	}

	amcTransaction := admin.Group("/amc-transaction")
	{
		amcTransaction.GET("", rbac.Check("amc", "view_amc_transactions"), amcTransactionHandler.GetAll)
		amcTransaction.GET("/:id", rbac.Check("amc", "view_amc_transactions_details"), amcTransactionHandler.GetByID)
	}

	amcOrder := admin.Group("/amc-order")
	{
		amcOrder.GET("", rbac.Check("amc", "view_amc_orders"), amcOrderHandler.GetAll)
		amcOrder.GET("/export", rbac.Check("amc", "export"), amcOrderHandler.ExportToCSV)
		amcOrder.GET("/:id", rbac.Check("amc", "view_amc_order_details"), amcOrderHandler.GetByID)
		amcOrder.PATCH("/:id/status", rbac.Check("amc", "status"), amcOrderHandler.UpdateStatus)
	}

	userRefund := admin.Group("/user-refund")
	{
		userRefund.GET("", rbac.Check("paymentandtransactions", "view_user_refund"), refundHandler.GetAllRefunds)
		userRefund.GET("/:id", rbac.Check("paymentandtransactions", "view_user_refund_details"), refundHandler.GetRefundByID)
	}

	amcRefund := admin.Group("/amc-refund")
	{
		amcRefund.GET("", rbac.Check("amc", "view_amc_refund"), amcRefundHandler.GetAll)
		amcRefund.GET("/:id", rbac.Check("amc", "view_amc_refund_details"), amcRefundHandler.GetDetails)
		amcRefund.PUT("/:id/approve", rbac.Check("amc", "approve_/_reject_refund"), amcRefundHandler.Approve)
		amcRefund.PUT("/:id/reject", rbac.Check("amc", "approve_/_reject_refund"), amcRefundHandler.Reject)
		amcRefund.GET("/:id/check-status", rbac.Check("amc", "approve_/_reject_refund"), amcRefundHandler.CheckStatus)
		amcRefund.GET("/stats", rbac.Check("amc", "view_amc_refund"), amcRefundHandler.GetStats)
	}

	zones := admin.Group("/zones")
	{
		zones.POST("", zoneHandler.Create)
		zones.GET("", rbac.Check("zones", "view_zones"), zoneHandler.GetAll)
		zones.GET("/active", zoneHandler.GetActive)
		zones.GET("/active-states", zoneHandler.GetActiveStates)
		zones.PUT("/:id", zoneHandler.Update)
		zones.PATCH("/:id/toggle-status", rbac.Check("zones", "activate/deactivate_zone"), zoneHandler.ToggleStatus)
		zones.DELETE("/:id", zoneHandler.Delete)

	}

	vehicleBrands := admin.Group("/vehicle-brands")
	{
		vehicleBrands.GET("", vehicleBrandHandler.GetBrands)
		vehicleBrands.GET("/models", vehicleBrandHandler.GetModels)
	}

	panelPermission := admin.Group("/permission")
	{
		panelPermission.GET("", permissionHandler.GetAllPermissions)
		panelPermission.GET("/:id", permissionHandler.GetPermissionByID)
		panelPermission.POST("", permissionHandler.CreatePermission)
		panelPermission.PUT("/:id", permissionHandler.UpdatePermission)
		panelPermission.DELETE("/:id", permissionHandler.DeletePermission)
	}

	zoneMap := admin.Group("/zoneMap")
	{
		zoneMap.GET("/stats", zoneMapHandler.GetZoneStats)
		zoneMap.GET("/:zone/activation-team", zoneMapHandler.GetActivationTeam)
		zoneMap.GET("/:zone/activators/:activator/providers", zoneMapHandler.GetProvidersByActivator)
		zoneMap.GET("/my-providers", zoneMapHandler.GetMyProviders)
	}

	vehicle := admin.Group("/vehicle")
	{
		vehicle.GET("/brands", vehicleHandler.GetVehicleBrands)
		vehicle.GET("/services", vehicleHandler.GetVehicleServices)
	}
}
