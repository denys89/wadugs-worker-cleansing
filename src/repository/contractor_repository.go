package repository

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type contractorRepository struct {
	db *gorm.DB
}

// NewContractorRepository creates a new contractor repository
func NewContractorRepository(db *gorm.DB) ContractorRepository {
	return &contractorRepository{
		db: db,
	}
}

func (r *contractorRepository) GetByID(ctx context.Context, id int64) (*entity.Contractor, error) {
	var contractor entity.Contractor
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contractor).Error
	if err != nil {
		return nil, err
	}
	return &contractor, nil
}

func (r *contractorRepository) GetAll(ctx context.Context) (entity.Contractors, error) {
	var contractors entity.Contractors
	err := r.db.WithContext(ctx).Find(&contractors).Error
	if err != nil {
		return nil, err
	}
	return contractors, nil
}

func (r *contractorRepository) GetByStatus(ctx context.Context, status int8) (entity.Contractors, error) {
	var contractors entity.Contractors
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&contractors).Error
	if err != nil {
		return nil, err
	}
	return contractors, nil
}

func (r *contractorRepository) Delete(ctx context.Context, id int64) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Contractor{}).Error
	if err != nil {
		return err
	}
	return nil
}