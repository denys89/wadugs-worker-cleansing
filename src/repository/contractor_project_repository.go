package repository

import (
	"context"

	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

// ContractorProjectRepository defines methods for contractor_project data access
type ContractorProjectRepository interface {
	GetByProjectID(ctx context.Context, projectID int64) (*entity.ContractorProject, error)
	GetByContractorID(ctx context.Context, contractorID int64) (entity.ContractorProjects, error)
	HardDeleteByContractorID(ctx context.Context, contractorID int64) error
}

type contractorProjectRepository struct {
	db *gorm.DB
}

// NewContractorProjectRepository creates a new contractor project repository
func NewContractorProjectRepository(db *gorm.DB) ContractorProjectRepository {
	return &contractorProjectRepository{
		db: db,
	}
}

// GetByProjectID returns the contractor_project record for a given project ID
// This is used to find which contractor a project belongs to
func (r *contractorProjectRepository) GetByProjectID(ctx context.Context, projectID int64) (*entity.ContractorProject, error) {
	var contractorProject entity.ContractorProject
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).First(&contractorProject).Error
	if err != nil {
		return nil, err
	}
	return &contractorProject, nil
}

// GetByContractorID returns all contractor_project records for a given contractor ID
func (r *contractorProjectRepository) GetByContractorID(ctx context.Context, contractorID int64) (entity.ContractorProjects, error) {
	var contractorProjects entity.ContractorProjects
	err := r.db.WithContext(ctx).Where("contractor_id = ?", contractorID).Find(&contractorProjects).Error
	if err != nil {
		return nil, err
	}
	return contractorProjects, nil
}

// HardDeleteByContractorID permanently deletes all contractor_project records for a contractor
func (r *contractorProjectRepository) HardDeleteByContractorID(ctx context.Context, contractorID int64) error {
	return r.db.WithContext(ctx).Where("contractor_id = ?", contractorID).Delete(&entity.ContractorProject{}).Error
}
