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
	cleansingService := r.ResolveCleansingService(ctx)
	s3Service, err := r.ResolveS3Service(ctx)
	if err != nil {
		panic(err)
	}
	
	handler := handlers.NewMessageHandler(cleansingService, s3Service)
	
	defer func() {
		log.Info("shutting down gracefully")
		consumer.Stop()
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