package repository

import (
	"context"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	"gorm.io/gorm"
)

type siteRepository struct {
	db *gorm.DB
}

// NewSiteRepository creates a new site repository
func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{
		db: db,
	}
}

func (r *siteRepository) GetByID(ctx context.Context, id int64) (*entity.Site, error) {
	var site entity.Site
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *siteRepository) GetAll(ctx context.Context) (entity.Sites, error) {
	var sites entity.Sites
	err := r.db.WithContext(ctx).Find(&sites).Error
	if err != nil {
		return nil, err
	}
	return sites, nil
}

func (r *siteRepository) GetByProjectID(ctx context.Context, projectID int64) (entity.Sites, error) {
	var sites entity.Sites
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&sites).Error
	if err != nil {
		return nil, err
	}
	return sites, nil
}

func (r *siteRepository) GetByStatus(ctx context.Context, status int8) (entity.Sites, error) {
	var sites entity.Sites
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&sites).Error
	if err != nil {
		return nil, err
	}
	return sites, nil
}