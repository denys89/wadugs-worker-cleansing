package repository

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

func (r *documentRepository) GetByID(ctx context.Context, id int64) (*entity.Document, error) {
	var document entity.Document
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) GetAll(ctx context.Context) (entity.Documents, error) {
	var documents entity.Documents
	err := r.db.WithContext(ctx).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *documentRepository) GetByGroupID(ctx context.Context, groupID int64) (entity.Documents, error) {
	var documents entity.Documents
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *documentRepository) GetByStatus(ctx context.Context, status int8) (entity.Documents, error) {
	var documents entity.Documents
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}