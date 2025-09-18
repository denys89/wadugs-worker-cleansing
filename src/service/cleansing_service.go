package service

import (
	"context"
	"fmt"
	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
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
		s3Service S3Service
	}

	// NullCleansingService is a no-op implementation for testing
	NullCleansingService struct{}
)

// NewCleansingService creates a new cleansing service instance
func NewCleansingService(s3Service S3Service) CleansingService {
	return &CleansingServiceImpl{
		s3Service: s3Service,
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

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"contractor_id":  contractorID,
		"files_deleted":  deletedCount,
	}).Info("Successfully deleted contractor files")

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
		return result, err
	}

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"project_id":    projectID,
		"files_deleted": deletedCount,
	}).Info("Successfully deleted project files")

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

	result.Success = true
	result.FilesDeleted = deletedCount
	logger.WithFields(log.Fields{
		"site_id":       siteID,
		"files_deleted": deletedCount,
	}).Info("Successfully deleted site files")

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