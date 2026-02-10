package repository

import (
	"context"

	"github.com/denys89/wadugs-worker-cleansing/src/entity"
)

// ContractorRepository defines methods for contractor data access
type ContractorRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Contractor, error)
	GetAll(ctx context.Context) (entity.Contractors, error)
	GetByStatus(ctx context.Context, status int8) (entity.Contractors, error)
	Delete(ctx context.Context, id int64) error
}

// UserContractorRepository defines methods for user_contractor data access
type UserContractorRepository interface {
	HardDeleteByContractorID(ctx context.Context, contractorID int64) error
}

// ViewerContractorRepository defines methods for viewer_contractor data access
type ViewerContractorRepository interface {
	HardDeleteByContractorID(ctx context.Context, contractorID int64) error
}

// ProjectRepository defines methods for project data access
type ProjectRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Project, error)
	GetAll(ctx context.Context) (entity.Projects, error)
	GetByContractorID(ctx context.Context, contractorID int64) (entity.Projects, error)
	GetByStatus(ctx context.Context, status int8) (entity.Projects, error)
	UpdateProjectUsage(ctx context.Context, projectID int64, sizeDelta int64) error
	HardDelete(ctx context.Context, id int64) error
	HardDeleteByContractorID(ctx context.Context, contractorID int64) error
	CleanupProjectAssociations(ctx context.Context, projectID int64) error
}

// SiteRepository defines methods for site data access
type SiteRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Site, error)
	GetAll(ctx context.Context) (entity.Sites, error)
	GetByProjectID(ctx context.Context, projectID int64) (entity.Sites, error)
	GetByStatus(ctx context.Context, status int8) (entity.Sites, error)
	HardDelete(ctx context.Context, id int64) error
	HardDeleteByProjectID(ctx context.Context, projectID int64) error
}

// DocumentGroupRepository defines methods for document group data access
type DocumentGroupRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.DocumentGroup, error)
	GetAll(ctx context.Context) (entity.DocumentGroups, error)
	GetBySiteID(ctx context.Context, siteID int64) (entity.DocumentGroups, error)
	GetByStatus(ctx context.Context, status int8) (entity.DocumentGroups, error)
	GetByProgress(ctx context.Context, progress int8) (entity.DocumentGroups, error)
	HardDeleteBySiteID(ctx context.Context, siteID int64) error
}

// DocumentRepository defines methods for document data access
type DocumentRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Document, error)
	GetAll(ctx context.Context) (entity.Documents, error)
	GetByGroupID(ctx context.Context, groupID int64) (entity.Documents, error)
	GetByStatus(ctx context.Context, status int8) (entity.Documents, error)
	HardDeleteBySiteID(ctx context.Context, siteID int64) error
}

// FileRepository defines methods for file data access
type FileRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.File, error)
	GetAll(ctx context.Context) (entity.Files, error)
	GetByDocumentID(ctx context.Context, documentID int64) (entity.Files, error)
	GetByStatus(ctx context.Context, status int8) (entity.Files, error)
	HardDeleteBySiteID(ctx context.Context, siteID int64) error
}
