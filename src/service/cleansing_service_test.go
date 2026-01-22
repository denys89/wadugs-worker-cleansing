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

// Mock project repository for testing
type mockProjectRepository struct{}

func (m *mockProjectRepository) GetByID(ctx context.Context, id int64) (*entity.Project, error) {
	return &entity.Project{
		Id:           id,
		Name:         "Test Project",
		ContractorId: 1,
	}, nil
}

func (m *mockProjectRepository) GetAll(ctx context.Context) (entity.Projects, error) {
	return entity.Projects{}, nil
}

func (m *mockProjectRepository) GetByContractorID(ctx context.Context, contractorID int64) (entity.Projects, error) {
	return entity.Projects{}, nil
}

func (m *mockProjectRepository) GetByStatus(ctx context.Context, status int8) (entity.Projects, error) {
	return entity.Projects{}, nil
}

func (m *mockProjectRepository) UpdateProjectUsage(ctx context.Context, projectID int64, sizeDelta int64) error {
	return nil
}

func (m *mockProjectRepository) HardDelete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockProjectRepository) HardDeleteByContractorID(ctx context.Context, contractorID int64) error {
	return nil
}

// Mock site repository for testing
type mockSiteRepository struct{}

func (m *mockSiteRepository) GetByID(ctx context.Context, id int64) (*entity.Site, error) {
	return &entity.Site{
		Id:        id,
		Name:      "Test Site",
		ProjectId: 1,
	}, nil
}

func (m *mockSiteRepository) GetAll(ctx context.Context) (entity.Sites, error) {
	return entity.Sites{}, nil
}

func (m *mockSiteRepository) GetByProjectID(ctx context.Context, projectID int64) (entity.Sites, error) {
	return entity.Sites{}, nil
}

func (m *mockSiteRepository) GetByStatus(ctx context.Context, status int8) (entity.Sites, error) {
	return entity.Sites{}, nil
}

func (m *mockSiteRepository) HardDelete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSiteRepository) HardDeleteByProjectID(ctx context.Context, projectID int64) error {
	return nil
}

// Mock document group repository for testing
type mockDocumentGroupRepository struct{}

func (m *mockDocumentGroupRepository) GetByID(ctx context.Context, id int64) (*entity.DocumentGroup, error) {
	return &entity.DocumentGroup{Id: id}, nil
}

func (m *mockDocumentGroupRepository) GetAll(ctx context.Context) (entity.DocumentGroups, error) {
	return entity.DocumentGroups{}, nil
}

func (m *mockDocumentGroupRepository) GetBySiteID(ctx context.Context, siteID int64) (entity.DocumentGroups, error) {
	return entity.DocumentGroups{}, nil
}

func (m *mockDocumentGroupRepository) GetByStatus(ctx context.Context, status int8) (entity.DocumentGroups, error) {
	return entity.DocumentGroups{}, nil
}

func (m *mockDocumentGroupRepository) GetByProgress(ctx context.Context, progress int8) (entity.DocumentGroups, error) {
	return entity.DocumentGroups{}, nil
}

func (m *mockDocumentGroupRepository) HardDeleteBySiteID(ctx context.Context, siteID int64) error {
	return nil
}

// Mock document repository for testing
type mockDocumentRepository struct{}

func (m *mockDocumentRepository) GetByID(ctx context.Context, id int64) (*entity.Document, error) {
	return &entity.Document{Id: id}, nil
}

func (m *mockDocumentRepository) GetAll(ctx context.Context) (entity.Documents, error) {
	return entity.Documents{}, nil
}

func (m *mockDocumentRepository) GetByGroupID(ctx context.Context, groupID int64) (entity.Documents, error) {
	return entity.Documents{}, nil
}

func (m *mockDocumentRepository) GetByStatus(ctx context.Context, status int8) (entity.Documents, error) {
	return entity.Documents{}, nil
}

func (m *mockDocumentRepository) HardDeleteBySiteID(ctx context.Context, siteID int64) error {
	return nil
}

// Mock file repository for testing
type mockFileRepository struct{}

func (m *mockFileRepository) GetByID(ctx context.Context, id int64) (*entity.File, error) {
	return &entity.File{Id: id}, nil
}

func (m *mockFileRepository) GetAll(ctx context.Context) (entity.Files, error) {
	return entity.Files{}, nil
}

func (m *mockFileRepository) GetByDocumentID(ctx context.Context, documentID int64) (entity.Files, error) {
	return entity.Files{}, nil
}

func (m *mockFileRepository) GetByStatus(ctx context.Context, status int8) (entity.Files, error) {
	return entity.Files{}, nil
}

func (m *mockFileRepository) HardDeleteBySiteID(ctx context.Context, siteID int64) error {
	return nil
}

func TestCleansingService_ProcessCleansingMessage(t *testing.T) {
	s3Service := NewNullS3Service()
	contractorRepo := &mockContractorRepository{}
	projectRepo := &mockProjectRepository{}
	siteRepo := &mockSiteRepository{}
	docGroupRepo := &mockDocumentGroupRepository{}
	docRepo := &mockDocumentRepository{}
	fileRepo := &mockFileRepository{}
	service := NewCleansingService(s3Service, contractorRepo, projectRepo, siteRepo, docGroupRepo, docRepo, fileRepo)

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
		name    string
		message dto.CleansingMessage
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
	projectRepo := &mockProjectRepository{}
	siteRepo := &mockSiteRepository{}
	docGroupRepo := &mockDocumentGroupRepository{}
	docRepo := &mockDocumentRepository{}
	fileRepo := &mockFileRepository{}
	service := NewCleansingService(s3Service, contractorRepo, projectRepo, siteRepo, docGroupRepo, docRepo, fileRepo)

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
	projectRepo := &mockProjectRepository{}
	siteRepo := &mockSiteRepository{}
	docGroupRepo := &mockDocumentGroupRepository{}
	docRepo := &mockDocumentRepository{}
	fileRepo := &mockFileRepository{}
	service := NewCleansingService(s3Service, contractorRepo, projectRepo, siteRepo, docGroupRepo, docRepo, fileRepo)
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
