package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/denys89/wadugs-worker-cleansing/src/dto"
	workerLog "github.com/denys89/wadugs-worker-cleansing/src/log"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

type (
	// S3Service defines the interface for S3 operations
	S3Service interface {
		ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error)
		ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error)
		ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error)
		DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error)
		DeleteBucket(ctx context.Context, bucketName string) error
	}

	// S3ServiceImpl implements the S3Service interface
	S3ServiceImpl struct {
		client          *s3.Client            // Default client for backward compatibility
		regionClients   map[string]*s3.Client // Cache of region-specific clients
		clientMutex     sync.RWMutex          // Mutex for thread-safe client cache access
		awsConfig       aws.Config            // AWS config for creating new clients
		accessKeyID     string                // AWS credentials
		secretAccessKey string
		rateLimiter     *rate.Limiter
		fileService     FileService
	}

	// NullS3Service is a no-op implementation for testing
	NullS3Service struct{}
)

const (
	// S3 batch delete limit (AWS maximum is 1000)
	maxDeleteBatchSize = 1000
	// Maximum concurrent delete operations (reduced for better rate limiting)
	maxConcurrentDeletes = 3
	// Rate limiting: 100 requests per second with burst of 10
	// This is conservative to avoid throttling
	requestsPerSecond = 100
	burstLimit        = 10
	// Retry configuration
	maxRetries = 3
	baseDelay  = 100 * time.Millisecond
	maxDelay   = 5 * time.Second
)

// NewS3Service creates a new S3 service instance with multi-region support
func NewS3Service(client *s3.Client, awsConfig aws.Config, accessKeyID, secretAccessKey string, fileService FileService) S3Service {
	// Create rate limiter: 100 requests per second with burst of 10
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burstLimit)

	return &S3ServiceImpl{
		client:          client,
		regionClients:   make(map[string]*s3.Client),
		awsConfig:       awsConfig,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		rateLimiter:     limiter,
		fileService:     fileService,
	}
}

// NewNullS3Service creates a null S3 service for testing
func NewNullS3Service() S3Service {
	return &NullS3Service{}
}

// getClientForRegion gets or creates an S3 client for a specific region
func (s3s *S3ServiceImpl) getClientForRegion(ctx context.Context, region string) (*s3.Client, error) {
	// If region is empty, use default client
	if region == "" {
		return s3s.client, nil
	}

	// Check if we already have a client for this region
	s3s.clientMutex.RLock()
	if client, exists := s3s.regionClients[region]; exists {
		s3s.clientMutex.RUnlock()
		return client, nil
	}
	s3s.clientMutex.RUnlock()

	// Create new client for this region
	s3s.clientMutex.Lock()
	defer s3s.clientMutex.Unlock()

	// Double-check in case another goroutine created it
	if client, exists := s3s.regionClients[region]; exists {
		return client, nil
	}

	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("region", region).Info("Creating new S3 client for region")

	// Create region-specific config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3s.accessKeyID,
			s3s.secretAccessKey,
			"", // session token
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config for region %s: %w", region, err)
	}

	// Create S3 client for this region
	client := s3.NewFromConfig(cfg)
	s3s.regionClients[region] = client

	logger.WithField("region", region).Info("Successfully created S3 client for region")
	return client, nil
}

// ListContractorFiles lists all S3 objects for a contractor across all buckets
func (s3s *S3ServiceImpl) ListContractorFiles(ctx context.Context, contractorID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("contractor_id", contractorID).Info("Listing contractor files")

	// Get file information from the file service
	objects, err := s3s.fileService.GetContractorFiles(ctx, contractorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contractor files: %w", err)
	}

	// TODO: Populate bucket information for each object
	// This would require additional logic to determine the correct bucket
	// based on project/site configuration or environment settings

	logger.WithFields(log.Fields{
		"contractor_id": contractorID,
		"total_files":   len(objects),
	}).Info("Listed contractor files")

	return objects, nil
}

// ListProjectFiles lists all S3 objects for a project
func (s3s *S3ServiceImpl) ListProjectFiles(ctx context.Context, projectID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("project_id", projectID).Info("Listing project files")

	// Get file information from the file service
	objects, err := s3s.fileService.GetProjectFiles(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project files: %w", err)
	}

	// TODO: Populate bucket information for each object
	// This would require additional logic to determine the correct bucket
	// based on project/site configuration or environment settings

	logger.WithFields(log.Fields{
		"project_id":  projectID,
		"total_files": len(objects),
	}).Info("Listed project files")

	return objects, nil
}

// ListSiteFiles lists all S3 objects for a site
func (s3s *S3ServiceImpl) ListSiteFiles(ctx context.Context, siteID int64) ([]dto.S3Object, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("site_id", siteID).Info("Listing site files")

	// Get file information from the file service
	objects, err := s3s.fileService.GetSiteFiles(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site files: %w", err)
	}

	// TODO: Populate bucket information for each object
	// This would require additional logic to determine the correct bucket
	// based on project/site configuration or environment settings

	logger.WithFields(log.Fields{
		"site_id":     siteID,
		"total_files": len(objects),
	}).Info("Listed site files")

	return objects, nil
}

// DeleteObjects deletes multiple S3 objects in batches with concurrency control and multi-region support
func (s3s *S3ServiceImpl) DeleteObjects(ctx context.Context, objects []dto.S3Object) (int, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("total_objects", len(objects)).Info("Starting multi-region batch delete operation")

	if len(objects) == 0 {
		return 0, nil
	}

	// Group objects by region and bucket for efficient batch deletion
	regionBucketObjects := make(map[string]map[string][]dto.S3Object)
	for _, obj := range objects {
		region := obj.Region
		if region == "" {
			logger.WithField("bucket", obj.Bucket).Warn("Object missing region, using default client")
		}

		if regionBucketObjects[region] == nil {
			regionBucketObjects[region] = make(map[string][]dto.S3Object)
		}
		regionBucketObjects[region][obj.Bucket] = append(regionBucketObjects[region][obj.Bucket], obj)
	}

	logger.WithField("regions_count", len(regionBucketObjects)).Info("Grouped objects by region")

	totalDeleted := 0
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrentDeletes)

	// Process each region
	for region, bucketObjects := range regionBucketObjects {
		region := region
		bucketObjects := bucketObjects

		// For each bucket in this region
		for bucket, bucketObjs := range bucketObjects {
			bucket := bucket
			bucketObjs := bucketObjs

			sem <- struct{}{}
			g.Go(func() error {
				defer func() { <-sem }()

				// Get region-specific client
				client, err := s3s.getClientForRegion(ctx, region)
				if err != nil {
					return fmt.Errorf("failed to get S3 client for region %s: %w", region, err)
				}

				deleted, err := s3s.deleteBucketObjectsWithClient(ctx, client, bucket, bucketObjs)
				if err != nil {
					return fmt.Errorf("failed to delete objects in bucket %s (region %s): %w", bucket, region, err)
				}
				totalDeleted += deleted
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return totalDeleted, err
	}

	logger.WithField("total_deleted", totalDeleted).Info("Completed multi-region batch delete operation")
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
			"bucket":        bucket,
			"batch_deleted": deleted,
			"total_deleted": totalDeleted,
		}).Debug("Deleted batch of objects")
	}

	return totalDeleted, nil
}

// deleteBucketObjectsWithClient deletes objects in a specific bucket using a specific S3 client
func (s3s *S3ServiceImpl) deleteBucketObjectsWithClient(ctx context.Context, client *s3.Client, bucket string, objects []dto.S3Object) (int, error) {
	logger := workerLog.GetLoggerFromContext(ctx)
	totalDeleted := 0

	// Process objects in batches
	for i := 0; i < len(objects); i += maxDeleteBatchSize {
		end := i + maxDeleteBatchSize
		if end > len(objects) {
			end = len(objects)
		}

		batch := objects[i:end]
		deleted, err := s3s.deleteBatchWithClient(ctx, client, bucket, batch)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket":     bucket,
				"batch_size": len(batch),
			}).Error("Failed to delete batch")
			return totalDeleted, err
		}

		totalDeleted += deleted
		logger.WithFields(log.Fields{
			"bucket":        bucket,
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

// deleteBatchWithClient deletes a batch of objects using a specific S3 client
func (s3s *S3ServiceImpl) deleteBatchWithClient(ctx context.Context, client *s3.Client, bucket string, objects []dto.S3Object) (int, error) {
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

	result, err := client.DeleteObjects(ctx, input)
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

// DeleteBucket deletes an S3 bucket after ensuring it's empty
// This implementation uses optimized batch operations, rate limiting, and retry logic
func (s3s *S3ServiceImpl) DeleteBucket(ctx context.Context, bucketName string) error {
	logger := workerLog.GetLoggerFromContext(ctx)
	logger.WithField("bucket", bucketName).Info("Starting optimized bucket deletion")

	// Step 1: Delete all objects in the bucket using optimized batch operations
	err := s3s.deleteAllObjectsInBucket(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to delete objects in bucket %s: %w", bucketName, err)
	}

	// Step 2: Delete the bucket itself with retry logic
	err = s3s.deleteBucketWithRetry(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", bucketName, err)
	}

	logger.WithField("bucket", bucketName).Info("Successfully deleted bucket")
	return nil
}

// deleteAllObjectsInBucket deletes all objects in a bucket using optimized pagination and batching
func (s3s *S3ServiceImpl) deleteAllObjectsInBucket(ctx context.Context, bucketName string) error {
	logger := workerLog.GetLoggerFromContext(ctx)
	totalDeleted := 0

	// Use paginated listing to handle large numbers of objects efficiently
	paginator := s3.NewListObjectsV2Paginator(s3s.client, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(1000), // Maximum page size for efficiency
	})

	// Process objects in batches as we paginate
	for paginator.HasMorePages() {
		// Rate limit the listing operation
		if err := s3s.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter context cancelled: %w", err)
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects page: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		// Convert to our DTO format
		var objects []dto.S3Object
		for _, obj := range page.Contents {
			objects = append(objects, dto.S3Object{
				Bucket: bucketName,
				Key:    aws.ToString(obj.Key),
			})
		}

		// Delete this batch of objects
		deleted, err := s3s.deleteBucketObjectsOptimized(ctx, bucketName, objects)
		if err != nil {
			return fmt.Errorf("failed to delete batch of %d objects: %w", len(objects), err)
		}

		totalDeleted += deleted
		logger.WithFields(log.Fields{
			"bucket":        bucketName,
			"batch_deleted": deleted,
			"total_deleted": totalDeleted,
		}).Info("Deleted batch of objects")
	}

	logger.WithFields(log.Fields{
		"bucket":        bucketName,
		"total_deleted": totalDeleted,
	}).Info("Completed deletion of all objects in bucket")

	return nil
}

// deleteBucketObjectsOptimized deletes objects with improved error handling and rate limiting
func (s3s *S3ServiceImpl) deleteBucketObjectsOptimized(ctx context.Context, bucket string, objects []dto.S3Object) (int, error) {
	if len(objects) == 0 {
		return 0, nil
	}

	logger := workerLog.GetLoggerFromContext(ctx)
	totalDeleted := 0

	// Process objects in batches of maxDeleteBatchSize (1000)
	for i := 0; i < len(objects); i += maxDeleteBatchSize {
		end := i + maxDeleteBatchSize
		if end > len(objects) {
			end = len(objects)
		}

		batch := objects[i:end]
		deleted, err := s3s.deleteBatchWithRetry(ctx, bucket, batch)
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"bucket":      bucket,
				"batch_size":  len(batch),
				"batch_start": i,
			}).Error("Failed to delete batch after retries")
			return totalDeleted, err
		}

		totalDeleted += deleted
		logger.WithFields(log.Fields{
			"bucket":        bucket,
			"batch_deleted": deleted,
			"total_deleted": totalDeleted,
			"progress":      fmt.Sprintf("%d/%d", end, len(objects)),
		}).Debug("Successfully deleted batch")
	}

	return totalDeleted, nil
}

// deleteBatchWithRetry implements exponential backoff retry for batch deletions
func (s3s *S3ServiceImpl) deleteBatchWithRetry(ctx context.Context, bucket string, objects []dto.S3Object) (int, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Rate limit each attempt
		if err := s3s.rateLimiter.Wait(ctx); err != nil {
			return 0, fmt.Errorf("rate limiter context cancelled: %w", err)
		}

		deleted, err := s3s.deleteBatch(ctx, bucket, objects)
		if err == nil {
			return deleted, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := time.Duration(attempt+1) * baseDelay
		if delay > maxDelay {
			delay = maxDelay
		}

		logger := workerLog.GetLoggerFromContext(ctx)
		logger.WithError(err).WithFields(log.Fields{
			"bucket":       bucket,
			"attempt":      attempt + 1,
			"max_attempts": maxRetries + 1,
			"retry_delay":  delay,
			"batch_size":   len(objects),
		}).Warn("Batch delete failed, retrying with backoff")

		// Wait before retry
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return 0, fmt.Errorf("batch delete failed after %d attempts: %w", maxRetries+1, lastErr)
}

// deleteBucketWithRetry deletes the bucket itself with retry logic
func (s3s *S3ServiceImpl) deleteBucketWithRetry(ctx context.Context, bucketName string) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Rate limit each attempt
		if err := s3s.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter context cancelled: %w", err)
		}

		_, err := s3s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := time.Duration(attempt+1) * baseDelay
		if delay > maxDelay {
			delay = maxDelay
		}

		logger := workerLog.GetLoggerFromContext(ctx)
		logger.WithError(err).WithFields(log.Fields{
			"bucket":       bucketName,
			"attempt":      attempt + 1,
			"max_attempts": maxRetries + 1,
			"retry_delay":  delay,
		}).Warn("Bucket deletion failed, retrying with backoff")

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("bucket deletion failed after %d attempts: %w", maxRetries+1, lastErr)
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

func (ns *NullS3Service) DeleteBucket(ctx context.Context, bucketName string) error {
	return nil
}
