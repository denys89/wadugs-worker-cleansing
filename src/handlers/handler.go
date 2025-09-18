package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
	"github.com/denys89/wadugs-worker-cleansing/src/service"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
	"time"
)

type (
	// MessageHandler handles incoming NSQ messages for cleansing operations
	MessageHandler struct {
		cleansingService service.CleansingService
		s3Service        service.S3Service
	}
)

// NewMessageHandler creates a new message handler instance
func NewMessageHandler(cleansingService service.CleansingService, s3Service service.S3Service) *MessageHandler {
	return &MessageHandler{
		cleansingService: cleansingService,
		s3Service:        s3Service,
	}
}

// HandleMessage processes incoming NSQ messages for cleansing operations
func (h *MessageHandler) HandleMessage(message *nsq.Message) error {
	// Create context with correlation ID for tracing
	correlationID := fmt.Sprintf("cleansing-%d", time.Now().UnixNano())
	ctx := context.Background()
	ctx = workerLog.WithLogger(ctx, correlationID)
	
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithFields(log.Fields{
		"message_id":      string(message.ID[:]),
		"message_body":    string(message.Body),
		"correlation_id":  correlationID,
		"attempts":        message.Attempts,
	}).Info("Received cleansing message")

	// Parse the message payload
	var cleansingMsg dto.CleansingMessage
	if err := json.Unmarshal(message.Body, &cleansingMsg); err != nil {
		logger.WithError(err).Error("Failed to unmarshal cleansing message")
		return h.handleError(ctx, fmt.Errorf("invalid message format: %w", err), false)
	}

	// Validate message type
	if !cleansingMsg.IsValidType() {
		logger.WithField("type", cleansingMsg.Type).Error("Invalid cleansing message type")
		return h.handleError(ctx, fmt.Errorf("invalid message type: %s", cleansingMsg.Type), false)
	}

	logger.WithFields(log.Fields{
		"type": cleansingMsg.Type,
		"id":   cleansingMsg.ID,
	}).Info("Processing cleansing request")

	// Process the cleansing operation
	result, err := h.processCleansingMessage(ctx, cleansingMsg)
	if err != nil {
		logger.WithError(err).Error("Failed to process cleansing message")
		return h.handleError(ctx, err, true) // Retry on processing errors
	}

	// Log the result
	logger.WithFields(log.Fields{
		"success":       result.Success,
		"files_deleted": result.FilesDeleted,
		"message":       result.Message,
	}).Info("Completed cleansing operation")

	return nil
}

// processCleansingMessage processes a cleansing message and returns the result
func (h *MessageHandler) processCleansingMessage(ctx context.Context, msg dto.CleansingMessage) (*dto.CleansingResult, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithFields(log.Fields{
		"type": msg.Type,
		"id":   msg.ID,
	}).Info("Processing cleansing message")

	// Validate message
	if !msg.IsValidType() {
		err := fmt.Errorf("invalid cleansing type: %s", msg.Type)
		logger.WithError(err).Error("Invalid message type")
		return nil, err
	}

	// Process the cleansing operation in the service layer
	result, err := h.cleansingService.ProcessCleansingMessage(ctx, msg)
	if err != nil {
		logger.WithError(err).Error("Failed to process cleansing in service layer")
		return result, err
	}

	logger.WithFields(log.Fields{
		"type":          msg.Type,
		"id":            msg.ID,
		"files_deleted": result.FilesDeleted,
	}).Info("Cleansing message processed successfully")

	return result, nil
}

// handleError handles errors during message processing
func (h *MessageHandler) handleError(ctx context.Context, err error, shouldRetry bool) error {
	logger := workerLog.GetLoggerFromContext(ctx)
	
	if shouldRetry {
		logger.WithError(err).Error("Retryable error occurred during message processing")
		return err // Return error to trigger NSQ retry
	}

	logger.WithError(err).Error("Non-retryable error occurred during message processing")
	// For non-retryable errors, we don't return the error to avoid infinite retries
	// The message will be marked as processed successfully but the error is logged
	return nil
}

// LogStats logs handler statistics (can be called periodically)
func (h *MessageHandler) LogStats() {
	log.WithFields(log.Fields{
		"handler": "cleansing",
		"status":  "active",
	}).Info("Message handler statistics")
}