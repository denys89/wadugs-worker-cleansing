package repository

import (
	"context"

	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type userContractorRepository struct {
	db *gorm.DB
}

// NewUserContractorRepository creates a new user_contractor repository
func NewUserContractorRepository(db *gorm.DB) UserContractorRepository {
	return &userContractorRepository{
		db: db,
	}
}

// HardDeleteByContractorID permanently deletes all user_contractor records for a contractor
func (r *userContractorRepository) HardDeleteByContractorID(ctx context.Context, contractorID int64) error {
	return r.db.WithContext(ctx).Where("contractor_id = ?", contractorID).Delete(&entity.UserContractor{}).Error
}
