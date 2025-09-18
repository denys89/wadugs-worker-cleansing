package log

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

// WithLogger adds a logger with correlation ID to the context
func WithLogger(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	logger := log.WithFields(log.Fields{
		"correlation_id": correlationID,
		"service":        "wadugs-worker-cleansing",
	})

	return context.WithValue(ctx, loggerKey, logger)
}

// GetLoggerFromContext retrieves the logger from context
func GetLoggerFromContext(ctx context.Context) *log.Entry {
	if logger, ok := ctx.Value(loggerKey).(*log.Entry); ok {
		return logger
	}

	// Return default logger if not found in context
	return log.WithFields(log.Fields{
		"service": "wadugs-worker-cleansing",
	})
}

// WithFields adds additional fields to the logger in context
func WithFields(ctx context.Context, fields log.Fields) context.Context {
	logger := GetLoggerFromContext(ctx)
	newLogger := logger.WithFields(fields)
	return context.WithValue(ctx, loggerKey, newLogger)
}