package repository

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) GetByID(ctx context.Context, id int64) (*entity.Project, error) {
	var project entity.Project
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetAll(ctx context.Context) (entity.Projects, error) {
	var projects entity.Projects
	err := r.db.WithContext(ctx).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByContractorID(ctx context.Context, contractorID int64) (entity.Projects, error) {
	var projects entity.Projects
	err := r.db.WithContext(ctx).Where("contractor_id = ?", contractorID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByStatus(ctx context.Context, status int8) (entity.Projects, error) {
	var projects entity.Projects
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}