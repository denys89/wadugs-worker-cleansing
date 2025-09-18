package repository

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		db: db,
	}
}

func (r *fileRepository) GetByID(ctx context.Context, id int64) (*entity.File, error) {
	var file entity.File
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetAll(ctx context.Context) (entity.Files, error) {
	var files entity.Files
	err := r.db.WithContext(ctx).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepository) GetByDocumentID(ctx context.Context, documentID int64) (entity.Files, error) {
	var files entity.Files
	err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepository) GetByStatus(ctx context.Context, status int8) (entity.Files, error) {
	var files entity.Files
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}