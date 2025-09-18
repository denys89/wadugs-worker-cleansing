package service

import (
	"context"
	"testing"
	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
)

// Mock contractor repository for testing
type mockContractorRepository struct{}

func (m *mockContractorRepository) GetByID(ctx context.Context, id int64) (*entity.Contractor, error) {
	return &entity.Contractor{
		Id:            id,
		Name:          "Test Contractor",
		AwsBucketName: "test-bucket",
	}, nil
}

func (m *mockContractorRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockContractorRepository) Create(ctx context.Context, contractor *entity.Contractor) error {
	return nil
}

func (m *mockContractorRepository) Update(ctx context.Context, contractor *entity.Contractor) error {
	return nil
}

func (m *mockContractorRepository) GetAll(ctx context.Context) (entity.Contractors, error) {
	return entity.Contractors{}, nil
}

func (m *mockContractorRepository) GetByStatus(ctx context.Context, status int8) (entity.Contractors, error) {
	return entity.Contractors{}, nil
}

func TestCleansingService_ProcessCleansingMessage(t *testing.T) {
	s3Service := NewNullS3Service()
	contractorRepo := &mockContractorRepository{}
	service := NewCleansingService(s3Service, contractorRepo)

	tests := []struct {
		name    string
		message dto.CleansingMessage
		wantErr bool
	}{
		{
			name:    "Valid contractor message",
			message: dto.CleansingMessage{Type: "contractor", ID: 123},
			wantErr: false,
		},
		{
			name:    "Valid project message",
			message: dto.CleansingMessage{Type: "project", ID: 456},
			wantErr: false,
		},
		{
			name:    "Valid site message",
			message: dto.CleansingMessage{Type: "site", ID: 789},
			wantErr: false,
		},
		{
			name:    "Invalid type",
			message: dto.CleansingMessage{Type: "invalid", ID: 123},
			wantErr: true,
		},
		{
			name:    "Empty type",
			message: dto.CleansingMessage{Type: "", ID: 123},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ProcessCleansingMessage(ctx, tt.message)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ProcessCleansingMessage() expected error but got none")
				}
				if result == nil || result.Success {
					t.Errorf("ProcessCleansingMessage() expected failed result")
				}
			} else {
				if err != nil {
					t.Errorf("ProcessCleansingMessage() unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Errorf("ProcessCleansingMessage() expected successful result")
				}
			}
		})
	}
}

func TestNullCleansingService_ProcessCleansingMessage(t *testing.T) {
	service := NewNullCleansingService()

	tests := []struct {
		name       string
		message    dto.CleansingMessage
	}{
		{
			name:    "Contractor message",
			message: dto.CleansingMessage{Type: "contractor", ID: 1},
		},
		{
			name:    "Project message",
			message: dto.CleansingMessage{Type: "project", ID: 2},
		},
		{
			name:    "Site message",
			message: dto.CleansingMessage{Type: "site", ID: 3},
		},
		{
			name:    "Invalid type should not error in null service",
			message: dto.CleansingMessage{Type: "invalid", ID: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ProcessCleansingMessage(ctx, tt.message)

			// Null service should never return an error
			if err != nil {
				t.Errorf("Null service should not return error, got: %v", err)
			}
			
			// Should return a successful result
			if result == nil || !result.Success {
				t.Error("Null service should return successful result")
			}
		})
	}
}

// Remove the old test methods that don't exist in the interface
func TestCleansingService_ValidEntityType(t *testing.T) {
	// This test is no longer needed as validation is done in the DTO
	t.Skip("Validation moved to DTO layer")
}

func TestCleansingService_ValidEntityID(t *testing.T) {
	// This test is no longer needed as validation is done in the DTO
	t.Skip("Validation moved to DTO layer")
}

func TestCleansingService_ErrorHandling(t *testing.T) {
	s3Service := NewNullS3Service()
	contractorRepo := &mockContractorRepository{}
	service := NewCleansingService(s3Service, contractorRepo)

	ctx := context.Background()

	// Test invalid message type
	invalidMessage := dto.CleansingMessage{Type: "invalid", ID: 123}
	result, err := service.ProcessCleansingMessage(ctx, invalidMessage)

	if err == nil {
		t.Error("Expected error for invalid message type")
	}

	if result == nil || result.Success {
		t.Error("Expected failed result for invalid message type")
	}
}

// Benchmark tests
func BenchmarkCleansingService_ProcessCleansingMessage(b *testing.B) {
	s3Service := NewNullS3Service()
	contractorRepo := &mockContractorRepository{}
	service := NewCleansingService(s3Service, contractorRepo)
	ctx := context.Background()
	message := dto.CleansingMessage{Type: "contractor", ID: 123}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ProcessCleansingMessage(ctx, message)
	}
}

func BenchmarkNullCleansingService_ProcessCleansingMessage(b *testing.B) {
	service := NewNullCleansingService()
	ctx := context.Background()
	message := dto.CleansingMessage{Type: "contractor", ID: 123}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ProcessCleansingMessage(ctx, message)
	}
}