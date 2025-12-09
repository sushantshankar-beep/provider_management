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
	"provider_management/internal/mq"
	"provider_management/internal/repository"
	"provider_management/internal/service"
)

func main() {
	cfg := config.Load()

	logg := logger.NewLogger()

	client := db.ConnectMongo(cfg.MongoURI)
	mongoDB := client.Database(cfg.MongoDBName)

	complaintRepo := repository.NewComplaintRepo(mongoDB)
	assessmentRepo := repository.NewAssessmentRepo(mongoDB)

	complaintService := service.NewComplaintService(complaintRepo, assessmentRepo)

	consumer := mq.NewConsumer(cfg.RabbitURL, complaintService, logger.NewLogger().Logger)
	go consumer.StartWorkers(5)

	h := handler.NewComplaintHandler(complaintService, logg)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h.Router(),
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
