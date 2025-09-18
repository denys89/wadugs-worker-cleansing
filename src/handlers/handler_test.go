package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	"github.com/nsqio/go-nsq"
)

// Mock services for testing
type mockCleansingService struct {
	shouldError   bool
	errorMsg      string
	filesDeleted  int
}

func (m *mockCleansingService) ProcessCleansingMessage(ctx context.Context, message dto.CleansingMessage) (*dto.CleansingResult, error) {
	if m.shouldError {
		return &dto.CleansingResult{
			Type:    message.Type,
			ID:      message.ID,
			Success: false,
			Error:   m.errorMsg,
		}, errors.New(m.errorMsg)
	}
	return &dto.CleansingResult{
		Type:         message.Type,
		ID:           message.ID,
		Success:      true,
		Message:      "Cleansing completed successfully",
		FilesDeleted: m.filesDeleted,
	}, nil
}

// Add the missing methods to match the CleansingService interface
func (m *mockCleansingService) DeleteContractorFiles(ctx context.Context, contractorID int64) (*dto.CleansingResult, error) {
	return m.ProcessCleansingMessage(ctx, dto.CleansingMessage{Type: "contractor", ID: contractorID})
}

func (m *mockCleansingService) DeleteProjectFiles(ctx context.Context, projectID int64) (*dto.CleansingResult, error) {
	return m.ProcessCleansingMessage(ctx, dto.CleansingMessage{Type: "project", ID: projectID})
}

func (m *mockCleansingService) DeleteSiteFiles(ctx context.Context, siteID int64) (*dto.CleansingResult, error) {
	return m.ProcessCleansingMessage(ctx, dto.CleansingMessage{Type: "site", ID: siteID})
}

type mockS3Service struct {
	shouldError   bool
	errorMsg      string
	filesToReturn []dto.S3Object
	deleteCount   int
}

func (m *mockS3Service) ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.filesToReturn, nil
}

func (m *mockS3Service) ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.filesToReturn, nil
}

func (m *mockS3Service) ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.filesToReturn, nil
}

func (m *mockS3Service) DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error) {
	if m.shouldError {
		return 0, errors.New(m.errorMsg)
	}
	return m.deleteCount, nil
}

// Add the missing DeleteBucket method to match the S3Service interface
func (m *mockS3Service) DeleteBucket(ctx context.Context, bucketName string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

func TestMessageHandler_HandleMessage_ValidMessages(t *testing.T) {
	tests := []struct {
		name    string
		message dto.CleansingMessage
	}{
		{
			name:    "Valid contractor message",
			message: dto.CleansingMessage{Type: "contractor", ID: 1},
		},
		{
			name:    "Valid project message",
			message: dto.CleansingMessage{Type: "project", ID: 2},
		},
		{
			name:    "Valid site message",
			message: dto.CleansingMessage{Type: "site", ID: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			cleansingService := &mockCleansingService{shouldError: false}
			s3Service := &mockS3Service{
				shouldError:   false,
				filesToReturn: []dto.S3Object{{Bucket: "test", Key: "test-key"}},
				deleteCount:   1,
			}

			handler := NewMessageHandler(cleansingService, s3Service)

			// Create NSQ message
			messageBody, _ := json.Marshal(tt.message)
			nsqMessage := &nsq.Message{
				Body: messageBody,
			}

			// Handle message
			err := handler.HandleMessage(nsqMessage)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestMessageHandler_HandleMessage_InvalidMessages(t *testing.T) {
	tests := []struct {
		name        string
		messageBody string
		expectError bool
	}{
		{
			name:        "Invalid JSON",
			messageBody: `{"type": "contractor", "id":}`,
			expectError: false, // Non-retryable error, returns nil
		},
		{
			name:        "Invalid message type",
			messageBody: `{"type": "invalid", "id": 1}`,
			expectError: false, // Non-retryable error, returns nil
		},
		{
			name:        "Missing type field",
			messageBody: `{"id": 1}`,
			expectError: false, // Non-retryable error, returns nil
		},
		{
			name:        "Missing id field",
			messageBody: `{"type": "contractor"}`,
			expectError: false, // Non-retryable error, returns nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			cleansingService := &mockCleansingService{shouldError: false}
			s3Service := &mockS3Service{shouldError: false}

			handler := NewMessageHandler(cleansingService, s3Service)

			// Create NSQ message
			nsqMessage := &nsq.Message{
				Body: []byte(tt.messageBody),
			}

			// Handle message
			err := handler.HandleMessage(nsqMessage)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestMessageHandler_HandleMessage_ServiceErrors(t *testing.T) {
	tests := []struct {
		name                     string
		message                  dto.CleansingMessage
		cleansingServiceError    bool
		cleansingServiceErrorMsg string
		expectRetryableError     bool
	}{
		{
			name:                     "Cleansing service error",
			message:                  dto.CleansingMessage{Type: "contractor", ID: 1},
			cleansingServiceError:    true,
			cleansingServiceErrorMsg: "Database connection failed",
			expectRetryableError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			cleansingService := &mockCleansingService{
				shouldError: tt.cleansingServiceError,
				errorMsg:    tt.cleansingServiceErrorMsg,
			}
			s3Service := &mockS3Service{
				shouldError:   false,
				filesToReturn: []dto.S3Object{{Bucket: "test", Key: "test-key"}},
				deleteCount:   1,
			}

			handler := NewMessageHandler(cleansingService, s3Service)

			// Create NSQ message
			messageBody, _ := json.Marshal(tt.message)
			nsqMessage := &nsq.Message{
				Body: messageBody,
			}

			// Handle message
			err := handler.HandleMessage(nsqMessage)
			if tt.expectRetryableError && err == nil {
				t.Errorf("Expected retryable error but got none")
			}
			if !tt.expectRetryableError && err != nil {
				t.Errorf("Expected no retryable error but got: %v", err)
			}
		})
	}
}

func TestMessageHandler_ProcessCleansingMessage_NoFiles(t *testing.T) {
	// Setup mocks with no files to delete
	cleansingService := &mockCleansingService{shouldError: false}
	s3Service := &mockS3Service{
		shouldError:   false,
		filesToReturn: []dto.S3Object{}, // No files
		deleteCount:   0,
	}

	handler := NewMessageHandler(cleansingService, s3Service)

	// Test message
	message := dto.CleansingMessage{Type: "contractor", ID: 1}

	result, err := handler.processCleansingMessage(context.Background(), message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success to be true")
	}

	if result.FilesDeleted != 0 {
		t.Errorf("Expected 0 files deleted, got %d", result.FilesDeleted)
	}
}

func TestMessageHandler_ProcessCleansingMessage_WithFiles(t *testing.T) {
	// Setup mocks with files to delete
	cleansingService := &mockCleansingService{
		shouldError:  false,
		filesDeleted: 2,
	}
	s3Service := &mockS3Service{
		shouldError: false,
		filesToReturn: []dto.S3Object{
			{Bucket: "test", Key: "test-key-1"},
			{Bucket: "test", Key: "test-key-2"},
		},
		deleteCount: 2,
	}

	handler := NewMessageHandler(cleansingService, s3Service)

	// Test message
	message := dto.CleansingMessage{Type: "project", ID: 2}

	result, err := handler.processCleansingMessage(context.Background(), message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success to be true")
	}

	if result.FilesDeleted != 2 {
		t.Errorf("Expected 2 files deleted, got %d", result.FilesDeleted)
	}
}

func TestMessageHandler_ProcessCleansingMessage_UnsupportedType(t *testing.T) {
	// Setup mocks
	cleansingService := &mockCleansingService{shouldError: false}
	s3Service := &mockS3Service{shouldError: false}

	handler := NewMessageHandler(cleansingService, s3Service)

	// Test message with unsupported type
	message := dto.CleansingMessage{Type: "unsupported", ID: 1}

	result, err := handler.processCleansingMessage(context.Background(), message)
	if err == nil {
		t.Errorf("Expected error for unsupported type")
	}

	if result != nil {
		t.Errorf("Expected nil result for unsupported type")
	}
}

func TestMessageHandler_HandleError(t *testing.T) {
	handler := &MessageHandler{}
	ctx := context.Background()
	testError := errors.New("test error")

	// Test retryable error
	err := handler.handleError(ctx, testError, true)
	if err == nil {
		t.Errorf("Expected error to be returned for retryable error")
	}

	// Test non-retryable error
	err = handler.handleError(ctx, testError, false)
	if err != nil {
		t.Errorf("Expected nil for non-retryable error, got: %v", err)
	}
}

// Benchmark tests
func BenchmarkMessageHandler_HandleMessage(b *testing.B) {
	// Setup mocks
	cleansingService := &mockCleansingService{shouldError: false}
	s3Service := &mockS3Service{
		shouldError:   false,
		filesToReturn: []dto.S3Object{{Bucket: "test", Key: "test-key"}},
		deleteCount:   1,
	}

	handler := NewMessageHandler(cleansingService, s3Service)

	// Create test message
	message := dto.CleansingMessage{Type: "contractor", ID: 1}
	messageBody, _ := json.Marshal(message)
	nsqMessage := &nsq.Message{
		Body: messageBody,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.HandleMessage(nsqMessage)
	}
}

func TestMessageHandler_LogStats(t *testing.T) {
	handler := &MessageHandler{}
	
	// This should not panic or error
	handler.LogStats()
}