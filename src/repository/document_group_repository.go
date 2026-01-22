package repository

import (
	"context"

	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type documentGroupRepository struct {
	db *gorm.DB
}

// NewDocumentGroupRepository creates a new document group repository
func NewDocumentGroupRepository(db *gorm.DB) DocumentGroupRepository {
	return &documentGroupRepository{
		db: db,
	}
}

func (r *documentGroupRepository) GetByID(ctx context.Context, id int64) (*entity.DocumentGroup, error) {
	var documentGroup entity.DocumentGroup
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&documentGroup).Error
	if err != nil {
		return nil, err
	}
	return &documentGroup, nil
}

func (r *documentGroupRepository) GetAll(ctx context.Context) (entity.DocumentGroups, error) {
	var documentGroups entity.DocumentGroups
	err := r.db.WithContext(ctx).Find(&documentGroups).Error
	if err != nil {
		return nil, err
	}
	return documentGroups, nil
}

func (r *documentGroupRepository) GetBySiteID(ctx context.Context, siteID int64) (entity.DocumentGroups, error) {
	var documentGroups entity.DocumentGroups
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).Find(&documentGroups).Error
	if err != nil {
		return nil, err
	}
	return documentGroups, nil
}

func (r *documentGroupRepository) GetByStatus(ctx context.Context, status int8) (entity.DocumentGroups, error) {
	var documentGroups entity.DocumentGroups
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&documentGroups).Error
	if err != nil {
		return nil, err
	}
	return documentGroups, nil
}

func (r *documentGroupRepository) GetByProgress(ctx context.Context, progress int8) (entity.DocumentGroups, error) {
	var documentGroups entity.DocumentGroups
	err := r.db.WithContext(ctx).Where("progress = ?", progress).Find(&documentGroups).Error
	if err != nil {
		return nil, err
	}
	return documentGroups, nil
}

// HardDeleteBySiteID permanently deletes all document groups belonging to a site
func (r *documentGroupRepository) HardDeleteBySiteID(ctx context.Context, siteID int64) error {
	return r.db.WithContext(ctx).Where("site_id = ?", siteID).Delete(&entity.DocumentGroup{}).Error
}
