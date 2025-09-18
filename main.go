package main

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/config"
	"github.com/denys89/wadugs-worker-cleansing/src/handlers"
	"github.com/denys89/wadugs-worker-cleansing/src/resolver"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	cfg := config.Get()

	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxAttempts = cfg.MaxRequeueAttempt
	consumer, err := nsq.NewConsumer(cfg.TopicName, cfg.ConsumerChannelName, nsqConfig)
	if err != nil {
		panic(err)
	}
	consumer.ChangeMaxInFlight(cfg.MaxInflight)
	
	// Create resolver and resolve services
	ctx := context.Background()
	r := resolver.NewResolver(cfg)
	
	// Initialize database connection
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database connection")
	}
	log.Info("Database connection established successfully")
	
	// Initialize repositories
	contractorRepo, err := r.ResolveContractorRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize contractor repository")
	}
	
	projectRepo, err := r.ResolveProjectRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize project repository")
	}
	
	siteRepo, err := r.ResolveSiteRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize site repository")
	}
	
	documentGroupRepo, err := r.ResolveDocumentGroupRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize document group repository")
	}
	
	documentRepo, err := r.ResolveDocumentRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize document repository")
	}
	
	fileRepo, err := r.ResolveFileRepository(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize file repository")
	}
	
	log.Info("All repositories initialized successfully")
	
	cleansingService := r.ResolveCleansingService(ctx)
	s3Service, err := r.ResolveS3Service(ctx)
	if err != nil {
		panic(err)
	}
	
	// Log repository availability for debugging
	log.WithFields(log.Fields{
		"contractor_repo":      contractorRepo != nil,
		"project_repo":         projectRepo != nil,
		"site_repo":           siteRepo != nil,
		"document_group_repo": documentGroupRepo != nil,
		"document_repo":       documentRepo != nil,
		"file_repo":           fileRepo != nil,
	}).Info("Repository status check")
	
	handler := handlers.NewMessageHandler(cleansingService, s3Service)
	
	defer func() {
		log.Info("shutting down gracefully")
		consumer.Stop()
		
		// Close database connection
		if db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				sqlDB.Close()
				log.Info("Database connection closed")
			}
		}
	}()
	
	consumer.AddConcurrentHandlers(
		handler,
		cfg.NsqConcurrency,
	)

	err = consumer.ConnectToNSQD(cfg.NsqServer)
	if err != nil {
		panic(err)
	}

	// Wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Info("received shutdown signal")
}