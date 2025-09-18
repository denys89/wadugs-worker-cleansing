package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	"github.com/denys89/wadugs-worker-cleansing/src/entity"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
	"github.com/denys89/wadugs-worker-cleansing/src/repository"
	log "github.com/sirupsen/logrus"
)

type (
	// FileService defines the interface for file operations and business logic
	FileService interface {
		GetContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error)
		GetProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error)
		GetSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error)
	}

	// FileServiceImpl implements the FileService interface
	FileServiceImpl struct {
		contractorRepo     repository.ContractorRepository
		projectRepo        repository.ProjectRepository
		siteRepo          repository.SiteRepository
		documentGroupRepo repository.DocumentGroupRepository
		documentRepo      repository.DocumentRepository
		fileRepo          repository.FileRepository
	}
)

// NewFileService creates a new file service instance
func NewFileService(
	contractorRepo repository.ContractorRepository,
	projectRepo repository.ProjectRepository,
	siteRepo repository.SiteRepository,
	documentGroupRepo repository.DocumentGroupRepository,
	documentRepo repository.DocumentRepository,
	fileRepo repository.FileRepository,
) FileService {
	return &FileServiceImpl{
		contractorRepo:    contractorRepo,
		projectRepo:       projectRepo,
		siteRepo:          siteRepo,
		documentGroupRepo: documentGroupRepo,
		documentRepo:      documentRepo,
		fileRepo:          fileRepo,
	}
}

// GetContractorFiles gets all file information for a contractor from the database
func (fs *FileServiceImpl) GetContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("contractor_id", contractorID).Info("Getting contractor files from database")

	var allObjects []dto.S3Object

	// 1. Query the database to get all projects for this contractor
	projects, err := fs.projectRepo.GetByContractorID(ctx, contractorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects for contractor %d: %w", contractorID, err)
	}

	logger.WithField("project_count", len(projects)).Info("Found projects for contractor")

	// Process each project
	for _, project := range projects {
		// 2. For each project, get all sites
		sites, err := fs.siteRepo.GetByProjectID(ctx, project.Id)
		if err != nil {
			logger.WithError(err).WithField("project_id", project.Id).Warn("Failed to get sites for project")
			continue
		}

		logger.WithFields(log.Fields{
			"project_id":  project.Id,
			"site_count":  len(sites),
		}).Debug("Found sites for project")

		// Process each site
		for _, site := range sites {
			// 3. For each site, get all document groups
			documentGroups, err := fs.documentGroupRepo.GetBySiteID(ctx, site.Id)
			if err != nil {
				logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to get document groups for site")
				continue
			}

			// Process each document group
			for _, docGroup := range documentGroups {
				// 4. For each document group, get all documents
				documents, err := fs.documentRepo.GetByGroupID(ctx, docGroup.Id)
				if err != nil {
					logger.WithError(err).WithField("group_id", docGroup.Id).Warn("Failed to get documents for group")
					continue
				}

				// Process each document
				for _, document := range documents {
					// 5. For each document, get all files
					files, err := fs.fileRepo.GetByDocumentID(ctx, document.Id)
					if err != nil {
						logger.WithError(err).WithField("document_id", document.Id).Warn("Failed to get files for document")
						continue
					}

					// 6. For each file, build S3 object information
					for _, file := range files {
						// Build S3 key based on the file path structure
						s3Objects := fs.buildS3ObjectsFromFile(project, site, docGroup, file)
						allObjects = append(allObjects, s3Objects...)
					}
				}

				// Handle processed files if they exist
				if (docGroup.Progress == 40 || docGroup.Progress == 11) && docGroup.ProcessedName != "" {
					processedObjects := fs.buildProcessedS3Objects(project, site, docGroup)
					allObjects = append(allObjects, processedObjects...)
				}
			}
		}
	}

	logger.WithFields(log.Fields{
		"contractor_id": contractorID,
		"total_files":   len(allObjects),
	}).Info("Retrieved contractor files from database")

	return allObjects, nil
}

// GetProjectFiles gets all file information for a project from the database
func (fs *FileServiceImpl) GetProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("project_id", projectID).Info("Getting project files from database")

	var allObjects []dto.S3Object

	// Get the project
	project, err := fs.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", projectID, err)
	}

	// Get all sites for this project
	sites, err := fs.siteRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sites for project %d: %w", projectID, err)
	}

	logger.WithField("site_count", len(sites)).Info("Found sites for project")

	// Process each site
	for _, site := range sites {
		// Get all document groups for this site
		documentGroups, err := fs.documentGroupRepo.GetBySiteID(ctx, site.Id)
		if err != nil {
			logger.WithError(err).WithField("site_id", site.Id).Warn("Failed to get document groups for site")
			continue
		}

		// Process each document group
		for _, docGroup := range documentGroups {
			// Get all documents for this group
			documents, err := fs.documentRepo.GetByGroupID(ctx, docGroup.Id)
			if err != nil {
				logger.WithError(err).WithField("group_id", docGroup.Id).Warn("Failed to get documents for group")
				continue
			}

			// Process each document
			for _, document := range documents {
				// Get all files for this document
				files, err := fs.fileRepo.GetByDocumentID(ctx, document.Id)
				if err != nil {
					logger.WithError(err).WithField("document_id", document.Id).Warn("Failed to get files for document")
					continue
				}

				// Build S3 objects from files
				for _, file := range files {
					s3Objects := fs.buildS3ObjectsFromFile(*project, site, docGroup, file)
					allObjects = append(allObjects, s3Objects...)
				}
			}

			// Handle processed files if they exist
			if (docGroup.Progress == 40 || docGroup.Progress == 11) && docGroup.ProcessedName != "" {
				processedObjects := fs.buildProcessedS3Objects(*project, site, docGroup)
				allObjects = append(allObjects, processedObjects...)
			}
		}
	}

	logger.WithFields(log.Fields{
		"project_id":  projectID,
		"total_files": len(allObjects),
	}).Info("Retrieved project files from database")

	return allObjects, nil
}

// GetSiteFiles gets all file information for a site from the database
func (fs *FileServiceImpl) GetSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("site_id", siteID).Info("Getting site files from database")

	var allObjects []dto.S3Object

	// Get the site
	site, err := fs.siteRepo.GetByID(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site %d: %w", siteID, err)
	}

	// Get the project for this site
	project, err := fs.projectRepo.GetByID(ctx, site.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d for site %d: %w", site.ProjectId, siteID, err)
	}

	// Get all document groups for this site
	documentGroups, err := fs.documentGroupRepo.GetBySiteID(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document groups for site %d: %w", siteID, err)
	}

	logger.WithField("document_group_count", len(documentGroups)).Info("Found document groups for site")

	// Process each document group
	for _, docGroup := range documentGroups {
		// Get all documents for this group
		documents, err := fs.documentRepo.GetByGroupID(ctx, docGroup.Id)
		if err != nil {
			logger.WithError(err).WithField("group_id", docGroup.Id).Warn("Failed to get documents for group")
			continue
		}

		// Process each document
		for _, document := range documents {
			// Get all files for this document
			files, err := fs.fileRepo.GetByDocumentID(ctx, document.Id)
			if err != nil {
				logger.WithError(err).WithField("document_id", document.Id).Warn("Failed to get files for document")
				continue
			}

			// Build S3 objects from files
			for _, file := range files {
				s3Objects := fs.buildS3ObjectsFromFile(*project, *site, docGroup, file)
				allObjects = append(allObjects, s3Objects...)
			}
		}

		// Handle processed files if they exist
		if (docGroup.Progress == 40 || docGroup.Progress == 11) && docGroup.ProcessedName != "" {
			processedObjects := fs.buildProcessedS3Objects(*project, *site, docGroup)
			allObjects = append(allObjects, processedObjects...)
		}
	}

	logger.WithFields(log.Fields{
		"site_id":     siteID,
		"total_files": len(allObjects),
	}).Info("Retrieved site files from database")

	return allObjects, nil
}

// buildS3ObjectsFromFile builds S3 object information from a file entity
func (fs *FileServiceImpl) buildS3ObjectsFromFile(project entity.Project, site entity.Site, docGroup entity.DocumentGroup, file entity.File) []dto.S3Object {
	var objects []dto.S3Object

	// Determine if we need to split the file name based on document group category
	var needSplit bool
	if docGroup.Category == "Boundary" ||
		docGroup.Category == "LineRoute" ||
		docGroup.Category == "SBEST" ||
		docGroup.Category == "SBP" ||
		docGroup.Category == "SoilSample" ||
		docGroup.Category == "SSS" {
		needSplit = true
	}

	// Build the base path: {projectCode}/{siteCode}/00_Upload/
	basePath := fmt.Sprintf("%s/%s/00_Upload/", project.Code, site.Code)
	
	fileName := file.Name
	if needSplit {
		fileNameSplit := strings.Split(file.Name, "/")
		if len(fileNameSplit) >= 2 {
			fileName = fmt.Sprintf("%s/Raw/%s", fileNameSplit[0], fileNameSplit[1])
		}
	}

	s3Key := fmt.Sprintf("%s%s", basePath, fileName)

	// Create S3 object with the key structure
	object := dto.S3Object{
		Key: s3Key,
		// Bucket would be populated based on project/site configuration
	}

	objects = append(objects, object)
	return objects
}

// buildProcessedS3Objects builds S3 objects for processed files
func (fs *FileServiceImpl) buildProcessedS3Objects(project entity.Project, site entity.Site, docGroup entity.DocumentGroup) []dto.S3Object {
	var objects []dto.S3Object

	// Build the processed files path: {projectCode}/{siteCode}/01_Processed/
	basePath := fmt.Sprintf("%s/%s/01_Processed/", project.Code, site.Code)

	// Add the main geojson file
	mainKey := fmt.Sprintf("%s%s.geojson", basePath, docGroup.ProcessedName)
	objects = append(objects, dto.S3Object{
		Key: mainKey,
	})

	// Add additional files for raster types
	if docGroup.Category == "RasterD" || docGroup.Category == "RasterO" || docGroup.Category == "Image" {
		fileTypes := []string{"_B01.tif", "_B02.tif", "_B03.tif"}
		for _, fileType := range fileTypes {
			key := fmt.Sprintf("%s%s%s", basePath, docGroup.ProcessedName, fileType)
			objects = append(objects, dto.S3Object{
				Key: key,
			})
		}
	}

	return objects
}