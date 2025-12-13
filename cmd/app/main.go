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

	complaintService := service.NewComplaintService(complaintRepo, assessmentRepo)
	transactionService := service.NewTransactionService(transactionRepo, acceptedServiceRepo, userRepo)

	complaintHandler := handler.NewComplaintHandler(complaintService, logg)
	transactionHandler := handler.NewTransactionHandler(transactionService, logg)

	r := gin.Default()

	r.GET("/complaints", complaintHandler.GetAll)
	r.GET("/complaints/:id", complaintHandler.GetByID)
	r.POST("/complaints/:id/assessment", complaintHandler.PostAssessment)

	r.GET("/transactions", transactionHandler.GetAll)
	r.GET("/transactions/:id", transactionHandler.GetByID)

	// consumer := mq.NewConsumer(cfg.RabbitURL, complaintService, logger.NewLogger().Logger)
	// go consumer.StartWorkers(5)

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