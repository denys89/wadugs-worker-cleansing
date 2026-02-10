package service

import (
	"context"
	"fmt"

	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
	"github.com/denys89/wadugs-worker-cleansing/src/repository"
	log "github.com/sirupsen/logrus"
)

type (
	// CleansingService defines the interface for data cleansing operations
	CleansingService interface {
		ProcessCleansingMessage(ctx context.Context, message dto.CleansingMessage) (*dto.CleansingResult, error)
		DeleteContractorFiles(ctx context.Context, contractorID int64) (*dto.CleansingResult, error)
		DeleteProjectFiles(ctx context.Context, projectID int64) (*dto.CleansingResult, error)
		DeleteSiteFiles(ctx context.Context, siteID int64) (*dto.CleansingResult, error)
	}

	// CleansingServiceImpl implements the CleansingService interface
	CleansingServiceImpl struct {
		s3Service         S3Service
		contractorRepo    repository.ContractorRepository
		projectRepo       repository.ProjectRepository
		siteRepo          repository.SiteRepository
		documentGroupRepo repository.DocumentGroupRepository
		documentRepo      repository.DocumentRepository
		fileRepo          repository.FileRepository
	}

	// NullCleansingService is a no-op implementation for testing
	NullCleansingService struct{}
)

// NewCleansingService creates a new cleansing service instance
func NewCleansingService(
	s3Service S3Service,
	contractorRepo repository.ContractorRepository,
	projectRepo repository.ProjectRepository,
	siteRepo repository.SiteRepository,
	documentGroupRepo repository.DocumentGroupRepository,
	documentRepo repository.DocumentRepository,
	fileRepo repository.FileRepository,
) CleansingService {
	return &CleansingServiceImpl{
		s3Service:         s3Service,
		contractorRepo:    contractorRepo,
		projectRepo:       projectRepo,
		siteRepo:          siteRepo,
		documentGroupRepo: documentGroupRepo,
		documentRepo:      documentRepo,
		fileRepo:          fileRepo,
	}
}

// NewNullCleansingService creates a null cleansing service for testing
func NewNullCleansingService() CleansingService {
	return &NullCleansingService{}
}

// ProcessCleansingMessage processes a cleansing message and routes to appropriate deletion method
func (cs *CleansingServiceImpl) ProcessCleansingMessage(ctx context.Context, message dto.CleansingMessage) (*dto.CleansingResult, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithFields(log.Fields{
		"type": message.Type,
		"id":   message.ID,
	}).Info("Processing cleansing message")

	if !message.IsValidType() {
		return &dto.CleansingResult{
			Type:    message.Type,
			ID:      message.ID,
			Success: false,
			Error:   fmt.Sprintf("invalid cleansing type: %s", message.Type),
		}, fmt.Errorf("invalid cleansing type: %s", message.Type)
	}

	switch message.Type {
	case dto.CleansingTypeContractor:
		return cs.DeleteContractorFiles(ctx, message.ID)
	case dto.CleansingTypeProject:
		return cs.DeleteProjectFiles(ctx, message.ID)
	case dto.CleansingTypeSite:
		return cs.DeleteSiteFiles(ctx, message.ID)
	default:
		return &dto.CleansingResult{
			Type:    message.Type,
			ID:      message.ID,
			Success: false,
			Error:   fmt.Sprintf("unsupported cleansing type: %s", message.Type),
		}, fmt.Errorf("unsupported cleansing type: %s", message.Type)
	}
}

// DeleteContractorFiles deletes all files related to a contractor (including all projects and sites)
func (cs *CleansingServiceImpl) DeleteContractorFiles(ctx context.Context, contractorID int64) (*dto.CleansingResult, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("contractor_id", contractorID).Info("Starting contractor file deletion")

	result := &dto.CleansingResult{
		Type:    dto.CleansingTypeContractor,
		ID:      contractorID,
		Success: false,
	}
	// Get all S3 objects for the contractor
	s3Objects, err := cs.s3Service.ListContractorFiles(ctx, contractorID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to list contractor files: %v", err)
		return result, err
	}

	logger.WithFields(log.Fields{
		"contractor_id": contractorID,
		"file_count":    len(s3Objects),
	}).Info("Found files to delete for contractor")

	// Delete all S3 objects
	deletedCount, err := cs.s3Service.DeleteObjects(ctx, s3Objects)
	if err != nil {
		result.Error = fmt.Sprintf("failed to delete contractor files: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// =====================================================
	// Database cascade deletion (bottom-up order)
	// =====================================================
	logger.WithField("contractor_id", contractorID).Info("Starting database cascade deletion for contractor")

	// Get all projects for this contractor to cascade delete their related records
	projects, err := cs.projectRepo.GetByContractorID(ctx, contractorID)
	if err != nil {
		logger.WithError(err).WithField("contractor_id", contractorID).Error("Failed to get projects for contractor")
		result.Error = fmt.Sprintf("failed to get projects for contractor: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// For each project, get all sites and cascade delete
	for _, project := range projects {
		sites, err := cs.siteRepo.GetByProjectID(ctx, project.Id)
		if err != nil {
			logger.WithError(err).WithField("project_id", project.Id).Warn("Failed to get sites for project during cascade")
			continue
		}

		for _, site := range sites {
			// 1. Delete all files belonging to documents of this site
			if err := cs.fileRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
				logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete file records for site")
			}

			// 2. Delete all documents belonging to document groups of this site
			if err := cs.documentRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
				logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete document records for site")
			}

			// 3. Delete all document groups of this site
			if err := cs.documentGroupRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
				logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete document group records for site")
			}
		}

		// 4. Delete all sites of this project
		if err := cs.siteRepo.HardDeleteByProjectID(ctx, project.Id); err != nil {
			logger.WithError(err).WithField("project_id", project.Id).Warn("Failed to delete site records for project")
		}
	}

	// 5. Delete all projects of this contractor
	if err := cs.projectRepo.HardDeleteByContractorID(ctx, contractorID); err != nil {
		logger.WithError(err).WithField("contractor_id", contractorID).Error("Failed to delete project records for contractor")
		result.Error = fmt.Sprintf("failed to delete project records: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// 6. Delete the contractor itself
	if err := cs.contractorRepo.Delete(ctx, contractorID); err != nil {
		logger.WithError(err).WithField("contractor_id", contractorID).Error("Failed to delete contractor record")
		result.Error = fmt.Sprintf("failed to delete contractor record: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"contractor_id": contractorID,
		"files_deleted": deletedCount,
	}).Info("Successfully deleted contractor files and database records")

	return result, nil
}

// DeleteProjectFiles deletes all files related to a project (including all sites)
func (cs *CleansingServiceImpl) DeleteProjectFiles(ctx context.Context, projectID int64) (*dto.CleansingResult, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("project_id", projectID).Info("Starting project file deletion")

	result := &dto.CleansingResult{
		Type:    dto.CleansingTypeProject,
		ID:      projectID,
		Success: false,
	}

	// Get all S3 objects for the project
	s3Objects, err := cs.s3Service.ListProjectFiles(ctx, projectID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to list project files: %v", err)
		return result, err
	}

	logger.WithFields(log.Fields{
		"project_id": projectID,
		"file_count": len(s3Objects),
	}).Info("Found files to delete for project")

	// Delete all S3 objects
	deletedCount, err := cs.s3Service.DeleteObjects(ctx, s3Objects)
	if err != nil {
		result.Error = fmt.Sprintf("failed to delete project files: %v", err)
		result.FilesDeleted = deletedCount

		// If some files were deleted before the error, still update usage for those
		if deletedCount > 0 {
			deletedSize := cs.calculateSizeForDeletedFiles(s3Objects, deletedCount)
			if updateErr := cs.projectRepo.UpdateProjectUsage(ctx, projectID, -deletedSize); updateErr != nil {
				logger.WithError(updateErr).WithFields(log.Fields{
					"project_id":   projectID,
					"deleted_size": deletedSize,
				}).Warn("Failed to update project usage for partially deleted files")
			} else {
				logger.WithFields(log.Fields{
					"project_id":    projectID,
					"files_deleted": deletedCount,
					"deleted_size":  deletedSize,
				}).Info("Updated project usage for partially deleted files")
			}
		}
		return result, err
	}

	// Calculate total file size for successfully deleted files
	var totalSize int64
	for _, obj := range s3Objects {
		totalSize += obj.Size
	}

	// Update usage metrics for successful deletion
	if err := cs.projectRepo.UpdateProjectUsage(ctx, projectID, -totalSize); err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"project_id": projectID,
			"total_size": totalSize,
		}).Error("Failed to update project usage after successful deletion")
		// Continue with database cleanup even if usage update fails
	}

	// =====================================================
	// Database cascade deletion (bottom-up order)
	// =====================================================
	logger.WithField("project_id", projectID).Info("Starting database cascade deletion for project")

	// Get all sites for this project to cascade delete
	sites, err := cs.siteRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		logger.WithError(err).WithField("project_id", projectID).Error("Failed to get sites for project")
		result.Error = fmt.Sprintf("failed to get sites for project: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	for _, site := range sites {
		// 1. Delete all files belonging to documents of this site
		if err := cs.fileRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
			logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete file records for site")
		}

		// 2. Delete all documents belonging to document groups of this site
		if err := cs.documentRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
			logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete document records for site")
		}

		// 3. Delete all document groups of this site
		if err := cs.documentGroupRepo.HardDeleteBySiteID(ctx, site.Id); err != nil {
			logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to delete document group records for site")
		}
	}

	// 4. Delete all sites of this project
	if err := cs.siteRepo.HardDeleteByProjectID(ctx, projectID); err != nil {
		logger.WithError(err).WithField("project_id", projectID).Error("Failed to delete site records for project")
		result.Error = fmt.Sprintf("failed to delete site records: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// 5. Delete the project itself
	if err := cs.projectRepo.HardDelete(ctx, projectID); err != nil {
		logger.WithError(err).WithField("project_id", projectID).Error("Failed to delete project record")
		result.Error = fmt.Sprintf("failed to delete project record: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"project_id":    projectID,
		"files_deleted": deletedCount,
		"total_size":    totalSize,
	}).Info("Successfully deleted project files and database records")

	return result, nil
}

// DeleteSiteFiles deletes all files related to a site
func (cs *CleansingServiceImpl) DeleteSiteFiles(ctx context.Context, siteID int64) (*dto.CleansingResult, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("site_id", siteID).Info("Starting site file deletion")

	result := &dto.CleansingResult{
		Type:    dto.CleansingTypeSite,
		ID:      siteID,
		Success: false,
	}

	// Get the site to obtain project ID for usage update
	site, err := cs.siteRepo.GetByID(ctx, siteID)
	if err != nil {
		logger.WithError(err).WithField("site_id", siteID).Error("Failed to get site information")
		result.Error = fmt.Sprintf("failed to get site information: %v", err)
		return result, err
	}

	// Get all S3 objects for the site
	s3Objects, err := cs.s3Service.ListSiteFiles(ctx, siteID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to list site files: %v", err)
		return result, err
	}

	logger.WithFields(log.Fields{
		"site_id":    siteID,
		"file_count": len(s3Objects),
	}).Info("Found files to delete for site")

	// Delete all S3 objects
	deletedCount, err := cs.s3Service.DeleteObjects(ctx, s3Objects)
	if err != nil {
		result.Error = fmt.Sprintf("failed to delete site files: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// Calculate total file size for successfully deleted files
	var totalSize int64
	for _, obj := range s3Objects {
		totalSize += obj.Size
	}

	// Update usage metrics for successful deletion
	if err := cs.projectRepo.UpdateProjectUsage(ctx, site.ProjectId, -totalSize); err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"site_id":    siteID,
			"project_id": site.ProjectId,
			"total_size": totalSize,
		}).Error("Failed to update project usage after successful site deletion")
		// Continue with database cleanup even if usage update fails
	}

	// =====================================================
	// Database cascade deletion (bottom-up order)
	// =====================================================
	logger.WithField("site_id", siteID).Info("Starting database cascade deletion for site")

	// 1. Delete all files belonging to documents of this site
	if err := cs.fileRepo.HardDeleteBySiteID(ctx, siteID); err != nil {
		logger.WithError(err).WithField("site_id", siteID).Error("Failed to delete file records")
		result.Error = fmt.Sprintf("failed to delete file records: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// 2. Delete all documents belonging to document groups of this site
	if err := cs.documentRepo.HardDeleteBySiteID(ctx, siteID); err != nil {
		logger.WithError(err).WithField("site_id", siteID).Error("Failed to delete document records")
		result.Error = fmt.Sprintf("failed to delete document records: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// 3. Delete all document groups of this site
	if err := cs.documentGroupRepo.HardDeleteBySiteID(ctx, siteID); err != nil {
		logger.WithError(err).WithField("site_id", siteID).Error("Failed to delete document group records")
		result.Error = fmt.Sprintf("failed to delete document group records: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	// 4. Delete the site itself
	if err := cs.siteRepo.HardDelete(ctx, siteID); err != nil {
		logger.WithError(err).WithField("site_id", siteID).Error("Failed to delete site record")
		result.Error = fmt.Sprintf("failed to delete site record: %v", err)
		result.FilesDeleted = deletedCount
		return result, err
	}

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"site_id":       siteID,
		"project_id":    site.ProjectId,
		"files_deleted": deletedCount,
		"total_size":    totalSize,
	}).Info("Successfully deleted site files and database records")

	return result, nil
}

// Null implementation methods for testing
func (ncs *NullCleansingService) ProcessCleansingMessage(ctx context.Context, message dto.CleansingMessage) (*dto.CleansingResult, error) {
	return &dto.CleansingResult{
		Type:         message.Type,
		ID:           message.ID,
		Success:      true,
		FilesDeleted: 0,
	}, nil
}

// calculateSizeForDeletedFiles calculates the total size of files that were successfully deleted
// This assumes files are deleted in order and the first 'deletedCount' files were successfully deleted
func (cs *CleansingServiceImpl) calculateSizeForDeletedFiles(s3Objects []dto.S3Object, deletedCount int) int64 {
	var totalSize int64

	// Only count the size of files that were actually deleted
	// The S3 service returns the count of successfully deleted files
	for i := 0; i < deletedCount && i < len(s3Objects); i++ {
		totalSize += s3Objects[i].Size
	}

	return totalSize
}

func (ncs *NullCleansingService) DeleteContractorFiles(ctx context.Context, contractorID int64) (*dto.CleansingResult, error) {
	return &dto.CleansingResult{
		Type:         dto.CleansingTypeContractor,
		ID:           contractorID,
		Success:      true,
		FilesDeleted: 0,
	}, nil
}

func (ncs *NullCleansingService) DeleteProjectFiles(ctx context.Context, projectID int64) (*dto.CleansingResult, error) {
	return &dto.CleansingResult{
		Type:         dto.CleansingTypeProject,
		ID:           projectID,
		Success:      true,
		FilesDeleted: 0,
	}, nil
}

func (ncs *NullCleansingService) DeleteSiteFiles(ctx context.Context, siteID int64) (*dto.CleansingResult, error) {
	return &dto.CleansingResult{
		Type:         dto.CleansingTypeSite,
		ID:           siteID,
		Success:      true,
		FilesDeleted: 0,
	}, nil
}
