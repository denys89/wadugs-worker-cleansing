package repository

import (
	"context"

	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type viewerContractorRepository struct {
	db *gorm.DB
}

// NewViewerContractorRepository creates a new viewer_contractor repository
func NewViewerContractorRepository(db *gorm.DB) ViewerContractorRepository {
	return &viewerContractorRepository{
		db: db,
	}
}

// HardDeleteByContractorID permanently deletes all viewer_contractor records for a contractor
func (r *viewerContractorRepository) HardDeleteByContractorID(ctx context.Context, contractorID int64) error {
	return r.db.WithContext(ctx).Where("contractor_id = ?", contractorID).Delete(&entity.ViewerContractor{}).Error
}
