package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"provider_management/internal/config"
	"provider_management/internal/db"
	"provider_management/internal/handler"
	"provider_management/internal/logger"
	"provider_management/internal/repository"
	"provider_management/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	logg := logger.NewLogger()

	client := db.ConnectMongo(cfg.MongoURI)
	mongoDB := client.Database(cfg.MongoDBName)

	complaintRepo := repository.NewComplaintRepo(mongoDB)
	assessmentRepo := repository.NewAssessmentRepo(mongoDB)
	transactionRepo := repository.NewTransactionRepo(mongoDB)
	acceptedServiceRepo := repository.NewAcceptedServiceRepo(mongoDB)
	userRepo := repository.NewUserRepo(mongoDB)
	vehiclesRepo := repository.NewSavedVehiclesRepo(mongoDB)
	amcRepo := repository.NewAMCPurchaseRepo(mongoDB)
	providerRepo := repository.NewProviderRepo(mongoDB)
	serviceRepo := repository.NewAcceptedServiceRepo(mongoDB)

	complaintService := service.NewComplaintService(complaintRepo, assessmentRepo)
	transactionService := service.NewTransactionService(transactionRepo, acceptedServiceRepo, userRepo)
	userAdminService := service.NewUserAdminService(userRepo, vehiclesRepo, acceptedServiceRepo, amcRepo)
	providerAdminService := service.NewProviderAdminService(providerRepo, serviceRepo)

	complaintHandler := handler.NewComplaintHandler(complaintService, logg)
	transactionHandler := handler.NewTransactionHandler(transactionService, logg)
	userAdminHandler := handler.NewUserAdminHandler(userAdminService)
    providerAdminHandler := handler.NewProviderAdminHandler(providerAdminService)

	
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:8002",
			"http://localhost:5173",
		},
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

	// consumer := mq.NewConsumer(cfg.RabbitURL, complaintService, logger.NewLogger().Logger)
	// go consumer.StartWorkers(5)
	r.GET("/admin/providers", providerAdminHandler.GetAll)
	r.GET("/admin/providers/:id", providerAdminHandler.GetByID)
	r.PATCH("/admin/providers/status/:id", providerAdminHandler.UpdateStatus)
	r.PATCH("/admin/providers/kyc/:id", providerAdminHandler.UpdateKYC)
	r.PATCH("/admin/providers/verify-document/:id", providerAdminHandler.VerifyDocument)
	r.PATCH("/admin/providers/:id/account-action", providerAdminHandler.UpdateAccountAction)
	r.PATCH("/admin/providers/commission/:id", providerAdminHandler.UpdateCommission)

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

