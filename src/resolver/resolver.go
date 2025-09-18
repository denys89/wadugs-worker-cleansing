package resolver

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	workerConfig "github.com/denys89/wadugs-worker-cleansing/src/config"
	"github.com/denys89/wadugs-worker-cleansing/src/service"
	log "github.com/sirupsen/logrus"
)

type (
	// Resolver handles dependency injection and service initialization
	Resolver struct {
		config *workerConfig.Config
	}
)

// NewResolver creates a new resolver instance
func NewResolver(cfg *workerConfig.Config) *Resolver {
	return &Resolver{
		config: cfg,
	}
}

// ResolveS3Client creates and configures an S3 client
func (r *Resolver) ResolveS3Client(ctx context.Context) (*s3.Client, error) {
	log.Info("Initializing S3 client")

	// Load AWS configuration
	var cfg aws.Config
	var err error

	// Check if we have explicit AWS credentials in config
	if r.config.AWSAccessKeyID != "" && r.config.AWSSecretAccessKey != "" {
		log.Info("Using explicit AWS credentials from configuration")
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(r.config.AWSRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				r.config.AWSAccessKeyID,
				r.config.AWSSecretAccessKey,
				"", // session token
			)),
		)
	} else {
		log.Info("Using default AWS credential chain")
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(r.config.AWSRegion),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Test the connection by listing buckets (optional health check)
	_, err = s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.WithError(err).Warn("Failed to test S3 connection - continuing anyway")
	} else {
		log.Info("S3 client initialized and tested successfully")
	}

	return s3Client, nil
}

// ResolveS3Service creates an S3 service instance
func (r *Resolver) ResolveS3Service(ctx context.Context) (service.S3Service, error) {
	log.Info("Resolving S3 service")

	// Create S3 client
	s3Client, err := r.ResolveS3Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve S3 client: %w", err)
	}

	// Create and return S3 service
	s3Service := service.NewS3Service(s3Client)
	log.Info("S3 service resolved successfully")

	return s3Service, nil
}

// ResolveCleansingService creates a cleansing service instance
func (r *Resolver) ResolveCleansingService(ctx context.Context) service.CleansingService {
	log.Info("Resolving cleansing service")

	// Create S3 service
	s3Service, err := r.ResolveS3Service(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve S3 service, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create and return cleansing service
	cleansingService := service.NewCleansingService(s3Service)
	log.Info("Cleansing service resolved successfully")

	return cleansingService
}

// ResolveAllServices resolves all required services for the application
func (r *Resolver) ResolveAllServices(ctx context.Context) (service.CleansingService, service.S3Service, error) {
	log.Info("Resolving all services")

	// Resolve cleansing service
	cleansingService := r.ResolveCleansingService(ctx)

	// Resolve S3 service
	s3Service, err := r.ResolveS3Service(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve S3 service: %w", err)
	}

	log.Info("All services resolved successfully")
	return cleansingService, s3Service, nil
}

// ValidateConfiguration validates the resolver configuration
func (r *Resolver) ValidateConfiguration() error {
	log.Info("Validating resolver configuration")

	if r.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate AWS configuration
	if r.config.AWSRegion == "" {
		return fmt.Errorf("AWS region is required")
	}

	if r.config.AWSAccessKeyID == "" {
		return fmt.Errorf("AWS access key ID is required")
	}

	if r.config.AWSSecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key is required")
	}

	// Validate NSQ configuration
	if r.config.NsqServer == "" {
		return fmt.Errorf("NSQ server is required")
	}

	if r.config.TopicName == "" {
		return fmt.Errorf("NSQ topic is required")
	}

	if r.config.ConsumerChannelName == "" {
		return fmt.Errorf("NSQ channel is required")
	}

	log.Info("Configuration validation completed successfully")
	return nil
}

// GetConfig returns the resolver configuration
func (r *Resolver) GetConfig() *workerConfig.Config {
	return r.config
}