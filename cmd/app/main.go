package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"os/signal"
	"provider_management/internal/config"
	"provider_management/internal/db"
	"provider_management/internal/handler"
	"provider_management/internal/middleware"
	"provider_management/internal/repository"
	"provider_management/internal/routes"
	"provider_management/internal/service"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env not loaded:", err)
	}
	cfg := config.Load()

	awsSession, err := config.InitAWSSession()
	if err != nil {
		log.Fatal("Failed to initialize AWS session:", err)
	}
	log.Println("AWS session initialized successfully")
	

	s3Uploader := middleware.NewS3Uploader(awsSession, os.Getenv("AWS_BUCKET_NAME"))

	client := db.ConnectMongo(cfg.MongoURI)
	mongoDB := client.Database(cfg.MongoDBName)



	complaintRepo := repository.NewComplaintRepository(mongoDB)
	transactionRepo := repository.NewTransactionRepo(mongoDB)
	acceptedServiceRepo := repository.NewAcceptedServiceRepo(mongoDB)
	userRepo := repository.NewUserRepo(mongoDB)
	vehiclesRepo := repository.NewSavedVehiclesRepo(mongoDB)
	amcRepo := repository.NewAMCPurchaseRepo(mongoDB)
	providerRepo := repository.NewProviderRepo(mongoDB)
	serviceRepo := repository.NewAcceptedServiceRepo(mongoDB)
	adminBookingRepo := repository.NewAdminBookingRepo(mongoDB)
	paymentPayoutRepo := repository.NewPaymentPayoutRepo(mongoDB)
	settlementRepo := repository.NewProviderSettlementRepo(mongoDB)
	serviceMasterRepo := repository.NewServiceMasterRepo(mongoDB)
	roleRepo := repository.NewRoleRepository(mongoDB)
	adminRepo := repository.NewAdminRepository(mongoDB)
	refundRepo := repository.NewRefundRepository(mongoDB)
	activityLogRepo := repository.NewActivityLogRepository(mongoDB)
	authMiddleware := middleware.NewAuthMiddleware(adminRepo)
	activityLogMiddleware := middleware.NewActivityLogMiddleware(activityLogRepo)
	rbacMiddleware := middleware.NewRBACMiddleware(roleRepo)
	amcPlanRepo := repository.NewAMCPlanRepo(mongoDB)
	amcTransactionRepo := repository.NewAMCTransactionRepo(mongoDB)
	amcOrderRepo := repository.NewOrderRepo(mongoDB)
	savedVehicleRepo := repository.NewSavedVehiclesRepo(mongoDB)
	zoneRepo := repository.NewZoneRepo(mongoDB)
	amcPurchaseRepo := repository.NewAMCPurchaseRepo(mongoDB)
	amcRefundRepo := repository.NewAMCRefundRepo(mongoDB)
	vehicleBrandRepo := repository.NewVehicleBrandRepo(mongoDB)
	bidRepo := repository.NewBidRepo(mongoDB)
	zoneFilterMiddleware := middleware.NewZoneFilterMiddleware(mongoDB)
	permissionRepo := repository. NewPermissionRepo(mongoDB)
	settlementHistoryRepo := repository.NewSettlementHistoryRepository(mongoDB)
    serviceR := repository.NewServiceRequestRepo(mongoDB)
	providerAgreementRepo := repository.NewAgreementRepo(mongoDB)
	kycRepo := repository.NewProviderKYCRepo(mongoDB)
	providerVehicleBrandRepo := repository.NewProviderVehicleBrandRepo(mongoDB)
    invoiceRepo := repository.NewInvoiceRepo(mongoDB)
	ratingRepo := repository.NewRatingRepo(mongoDB)

	transactionService := service.NewTransactionService(transactionRepo, acceptedServiceRepo, userRepo)
	userAdminService := service.NewUserAdminService(userRepo, vehiclesRepo, acceptedServiceRepo, amcRepo)
	providerAdminService := service.NewProviderAdminService(providerRepo, serviceRepo,adminRepo,zoneRepo,roleRepo, settlementRepo, settlementHistoryRepo, serviceR,kycRepo)
	adminBookingService := service.NewAdminBookingService(adminBookingRepo,invoiceRepo,transactionRepo,settlementHistoryRepo,ratingRepo)
	payoutService := service.NewPayoutService(acceptedServiceRepo, paymentPayoutRepo, providerRepo, settlementRepo,kycRepo,transactionRepo)
	settlementService := service.NewSettlementService(serviceRepo, settlementRepo, paymentPayoutRepo, providerRepo,settlementHistoryRepo,kycRepo,transactionRepo)
	refundService := service.NewRefundService(refundRepo, transactionRepo, userRepo)
	complaintService := service.NewComplaintService(paymentPayoutRepo,complaintRepo, acceptedServiceRepo, userRepo, providerRepo, refundService, payoutService,transactionRepo,kycRepo)
	serviceMasterService := service.NewServiceMaster(serviceMasterRepo)
	adminService := service.NewAdminService(adminRepo, roleRepo)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	amcPlanService := service.NewAMCPlanService(amcPlanRepo)
	amcTransactionService := service.NewAMCTransactionService(amcTransactionRepo, userRepo, amcRepo)
	amcOrderService := service.NewOrderService(amcOrderRepo, userRepo, amcPlanRepo, savedVehicleRepo, zoneRepo)
	payUService := service.NewPayUService(cfg.PayU.Key, cfg.PayU.Salt, cfg.PayU.BaseURL)
	amcRefundService := service.NewAMCRefundService(amcRefundRepo, amcPurchaseRepo, userRepo, amcPlanRepo, payUService)
	adminRoleService := service.NewRoleService(roleRepo)
	dashboardService := service.NewDashboardService(providerRepo, userRepo, acceptedServiceRepo, settlementRepo, complaintRepo,transactionRepo, amcPurchaseRepo, bidRepo)   
	permissionService := service.NewPermissionService(permissionRepo)
    zoneMapService := service.NewZoneMapService(providerRepo,adminRepo,roleRepo,acceptedServiceRepo)
	providerAgreementService := service.NewAgreementService(providerAgreementRepo)
	providerBrandService := service.NewProviderBrandService(providerVehicleBrandRepo,serviceMasterRepo)
	 
	complaintHandler := handler.NewComplaintHandler(complaintService, acceptedServiceRepo)
	zoneService := service.NewZoneService(zoneRepo)
	vehicleBrandService := service.NewVehicleBrandService(vehicleBrandRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	userAdminHandler := handler.NewUserAdminHandler(userAdminService)
	providerAdminHandler := handler.NewProviderAdminHandler(providerAdminService)
	adminBookingHandler := handler.NewAdminBookingHandler(adminBookingService)
	payoutHandler := handler.NewPayoutHandler(payoutService)
	settlementHandler := handler.NewSettlementHandler(settlementService)
	serviceMasterHandler := handler.NewServiceHandler(serviceMasterService)
	adminHandler := handler.NewAdminHandler(adminService)
	activityLogHandler := handler.NewActivityLogHandler(activityLogService)
	amcPlanHandler := handler.NewAMCPlanHandler(amcPlanService)
	amcTransactionHandler := handler.NewAMCTransactionHandler(amcTransactionService)
	amcOrderHandler := handler.NewOrderHandler(amcOrderService)
	refundHandler := handler.NewRefundHandler(refundService)
	amcRefundHandler := handler.NewAMCRefundHandler(amcRefundService)
	roleHandler := handler.NewRoleHandler(roleRepo, adminRoleService, adminRepo)
	zoneHandler := handler.NewZoneHandler(zoneService)
	vehicleBrandHandler := handler.NewVehicleBrandHandler(vehicleBrandService)
    dashboardHandler := handler.NewDashboardHandler(dashboardService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
    zoneMapHandler := handler.NewZoneMapHandler(zoneMapService)
	providerAgreementHandler := handler.NewAgreementHandler(providerAgreementService)
    providerBrandServiceHandler := handler.NewProviderBrandServiceHandler(providerBrandService)

	r := gin.Default() 
	r.SetTrustedProxies(nil)
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	routes.SetupRoutes(
		r,
		cfg.AllowedOrigins,
		complaintHandler,
		transactionHandler,
		userAdminHandler,
		providerAdminHandler,
		adminBookingHandler,
		payoutHandler,
		settlementHandler,
		serviceMasterHandler,
		adminHandler,
		activityLogHandler,
		amcPlanHandler,
		amcTransactionHandler,
		amcOrderHandler,
		refundHandler,
		amcRefundHandler,
		authMiddleware,
		activityLogMiddleware,
		rbacMiddleware,
		roleHandler,
		zoneHandler,
		vehicleBrandHandler,
		dashboardHandler,
		s3Uploader,
		zoneFilterMiddleware,
		permissionHandler,
		zoneMapHandler,
		providerBrandServiceHandler,
		providerAgreementHandler,
	)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}
	go func() {
		log.Println("🚀 Starting HTTP server on", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

