package repository

import (
	"context"
	"fmt"

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
	// Join with contractor_project table to find projects for this contractor
	err := r.db.WithContext(ctx).
		Table("project").
		Joins("INNER JOIN contractor_project ON contractor_project.project_id = project.id").
		Where("contractor_project.contractor_id = ?", contractorID).
		Find(&projects).Error
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

func (r *projectRepository) UpdateProjectUsage(ctx context.Context, projectID int64, sizeDelta int64) error {
	// Update project usage - this could be updating a usage field in the project table
	// For now, we'll implement a basic update that could be extended based on actual requirements
	err := r.db.WithContext(ctx).Model(&entity.Project{}).
		Where("id = ?", projectID).
		Update("updated_at", "NOW()").Error
	if err != nil {
		return err
	}
	return nil
}

// HardDelete permanently deletes a project by ID
func (r *projectRepository) HardDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Project{}, "id = ?", id).Error
}

// HardDeleteByContractorID permanently deletes all projects belonging to a contractor
func (r *projectRepository) HardDeleteByContractorID(ctx context.Context, contractorID int64) error {
	return r.db.WithContext(ctx).
		Exec("DELETE FROM project WHERE id IN (SELECT project_id FROM contractor_project WHERE contractor_id = ?)", contractorID).
		Error
}

// CleanupProjectAssociations deletes all FK-blocking association records for a project
// These tables have foreign keys referencing project(id) without ON DELETE CASCADE:
// client_project, uploader_project, vessel_project, contractor_project
func (r *projectRepository) CleanupProjectAssociations(ctx context.Context, projectID int64) error {
	// Delete from client_project
	if err := r.db.WithContext(ctx).Exec("DELETE FROM client_project WHERE project_id = ?", projectID).Error; err != nil {
		return fmt.Errorf("failed to delete client_project records: %w", err)
	}

	// Delete from uploader_project
	if err := r.db.WithContext(ctx).Exec("DELETE FROM uploader_project WHERE project_id = ?", projectID).Error; err != nil {
		return fmt.Errorf("failed to delete uploader_project records: %w", err)
	}

	// Delete from vessel_project
	if err := r.db.WithContext(ctx).Exec("DELETE FROM vessel_project WHERE project_id = ?", projectID).Error; err != nil {
		return fmt.Errorf("failed to delete vessel_project records: %w", err)
	}

	// Delete from contractor_project
	if err := r.db.WithContext(ctx).Exec("DELETE FROM contractor_project WHERE project_id = ?", projectID).Error; err != nil {
		return fmt.Errorf("failed to delete contractor_project records: %w", err)
	}

	return nil
}
