package service

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type (
	// S3Service defines the interface for S3 operations
	S3Service interface {
		ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error)
		ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error)
		ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error)
		DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error)
	}

	// S3ServiceImpl implements the S3Service interface
	S3ServiceImpl struct {
		client *s3.Client
	}

	// NullS3Service is a no-op implementation for testing
	NullS3Service struct{}
)

const (
	// S3 batch delete limit
	maxDeleteBatchSize = 1000
	// Maximum concurrent delete operations
	maxConcurrentDeletes = 5
)

// NewS3Service creates a new S3 service instance
func NewS3Service(client *s3.Client) S3Service {
	return &S3ServiceImpl{
		client: client,
	}
}

// NewNullS3Service creates a null S3 service for testing
func NewNullS3Service() S3Service {
	return &NullS3Service{}
}

// ListContractorFiles lists all S3 objects for a contractor across all buckets
func (s3s *S3ServiceImpl) ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("contractor_id", contractorID).Info("Listing contractor files")

	// In a real implementation, you would query the database to get all buckets
	// and prefixes associated with this contractor. For now, we'll use a pattern-based approach.
	
	// This is a simplified implementation. In practice, you would:
	// 1. Query the database to get all projects for this contractor
	// 2. For each project, get all sites
	// 3. For each site, get the bucket and prefix information
	// 4. List objects in S3 using those prefixes
	
	var allObjects []dto.S3Object
	
	// For demonstration, we'll assume a naming convention where contractor files
	// are stored with a prefix pattern. In reality, you'd query the database.
	buckets, err := s3s.listAllBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	for _, bucket := range buckets {
		// Look for objects with contractor-related prefixes
		// This is a simplified approach - in practice, you'd use database queries
		prefix := fmt.Sprintf("contractor_%d/", contractorID)
		objects, err := s3s.listObjectsWithPrefix(ctx, bucket, prefix)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket": bucket,
				"prefix": prefix,
			}).Warn("Failed to list objects for contractor prefix")
			continue
		}
		allObjects = append(allObjects, objects...)
	}

	logger.WithFields(log.Fields{
		"contractor_id": contractorID,
		"total_files":   len(allObjects),
	}).Info("Listed contractor files")

	return allObjects, nil
}

// ListProjectFiles lists all S3 objects for a project
func (s3s *S3ServiceImpl) ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("project_id", projectID).Info("Listing project files")

	var allObjects []dto.S3Object
	
	buckets, err := s3s.listAllBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	for _, bucket := range buckets {
		// Look for objects with project-related prefixes
		prefix := fmt.Sprintf("project_%d/", projectID)
		objects, err := s3s.listObjectsWithPrefix(ctx, bucket, prefix)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket": bucket,
				"prefix": prefix,
			}).Warn("Failed to list objects for project prefix")
			continue
		}
		allObjects = append(allObjects, objects...)
	}

	logger.WithFields(log.Fields{
		"project_id":  projectID,
		"total_files": len(allObjects),
	}).Info("Listed project files")

	return allObjects, nil
}

// ListSiteFiles lists all S3 objects for a site
func (s3s *S3ServiceImpl) ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("site_id", siteID).Info("Listing site files")

	var allObjects []dto.S3Object
	
	buckets, err := s3s.listAllBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	for _, bucket := range buckets {
		// Look for objects with site-related prefixes
		prefix := fmt.Sprintf("site_%d/", siteID)
		objects, err := s3s.listObjectsWithPrefix(ctx, bucket, prefix)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket": bucket,
				"prefix": prefix,
			}).Warn("Failed to list objects for site prefix")
			continue
		}
		allObjects = append(allObjects, objects...)
	}

	logger.WithFields(log.Fields{
		"site_id":     siteID,
		"total_files": len(allObjects),
	}).Info("Listed site files")

	return allObjects, nil
}

// DeleteObjects deletes multiple S3 objects in batches with concurrency control
func (s3s *S3ServiceImpl) DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("total_objects", len(objects)).Info("Starting batch delete operation")

	if len(objects) == 0 {
		return 0, nil
	}

	// Group objects by bucket for efficient batch deletion
	bucketObjects := make(map[string][]dto.S3Object)
	for _, obj := range objects {
		bucketObjects[obj.Bucket] = append(bucketObjects[obj.Bucket], obj)
	}

	totalDeleted := 0
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrentDeletes)

	for bucket, bucketObjs := range bucketObjects {
		bucket := bucket
		bucketObjs := bucketObjs

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()

			deleted, err := s3s.deleteBucketObjects(ctx, bucket, bucketObjs)
			if err != nil {
				return fmt.Errorf("failed to delete objects in bucket %s: %w", bucket, err)
			}
			totalDeleted += deleted
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return totalDeleted, err
	}

	logger.WithField("total_deleted", totalDeleted).Info("Completed batch delete operation")
	return totalDeleted, nil
}

// deleteBucketObjects deletes objects in a specific bucket using batch operations
func (s3s *S3ServiceImpl) deleteBucketObjects(ctx context.Context, bucket string, objects []dto.S3Object) (int, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	totalDeleted := 0

	// Process objects in batches
	for i := 0; i < len(objects); i += maxDeleteBatchSize {
		end := i + maxDeleteBatchSize
		if end > len(objects) {
			end = len(objects)
		}

		batch := objects[i:end]
		deleted, err := s3s.deleteBatch(ctx, bucket, batch)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket":     bucket,
				"batch_size": len(batch),
			}).Error("Failed to delete batch")
			return totalDeleted, err
		}

		totalDeleted += deleted
		logger.WithFields(log.Fields{
			"bucket":       bucket,
			"batch_deleted": deleted,
			"total_deleted": totalDeleted,
		}).Debug("Deleted batch of objects")
	}

	return totalDeleted, nil
}

// deleteBatch deletes a batch of objects using S3 batch delete API
func (s3s *S3ServiceImpl) deleteBatch(ctx context.Context, bucket string, objects []dto.S3Object) (int, error) {
	if len(objects) == 0 {
		return 0, nil
	}

	// Prepare delete objects
	var deleteObjects []types.ObjectIdentifier
	for _, obj := range objects {
		deleteObjects = append(deleteObjects, types.ObjectIdentifier{
			Key: aws.String(obj.Key),
		})
	}

	// Perform batch delete
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: deleteObjects,
			Quiet:   aws.Bool(false), // We want to see what was deleted
		},
	}

	result, err := s3s.client.DeleteObjects(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to delete objects: %w", err)
	}

	// Check for errors in the response
	if len(result.Errors) > 0 {
		logger := workerLog.GetLoggerFromContext(ctx)
		for _, deleteError := range result.Errors {
			logger.WithFields(log.Fields{
				"key":   aws.ToString(deleteError.Key),
				"code":  aws.ToString(deleteError.Code),
				"error": aws.ToString(deleteError.Message),
			}).Error("Failed to delete object")
		}
	}

	return len(result.Deleted), nil
}

// listAllBuckets lists all S3 buckets accessible to the service
func (s3s *S3ServiceImpl) listAllBuckets(ctx context.Context) ([]string, error) {
	result, err := s3s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var buckets []string
	for _, bucket := range result.Buckets {
		buckets = append(buckets, aws.ToString(bucket.Name))
	}

	return buckets, nil
}

// listObjectsWithPrefix lists all objects in a bucket with a specific prefix
func (s3s *S3ServiceImpl) listObjectsWithPrefix(ctx context.Context, bucket, prefix string) ([]dto.S3Object, error) {
	var objects []dto.S3Object
	
	paginator := s3.NewListObjectsV2Paginator(s3s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, dto.S3Object{
				Bucket: bucket,
				Key:    aws.ToString(obj.Key),
			})
		}
	}

	return objects, nil
}

// Null implementation methods for testing
func (ns *NullS3Service) ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error) {
	return []dto.S3Object{}, nil
}

func (ns *NullS3Service) ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error) {
	return []dto.S3Object{}, nil
}

func (ns *NullS3Service) ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error) {
	return []dto.S3Object{}, nil
}

func (ns *NullS3Service) DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error) {
	return len(objects), nil
}