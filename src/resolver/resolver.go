package resolver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	workerConfig "github.com/denys89/wadugs-worker-cleansing/src/config"
	"github.com/denys89/wadugs-worker-cleansing/src/database"
	"github.com/denys89/wadugs-worker-cleansing/src/repository"
	"github.com/denys89/wadugs-worker-cleansing/src/service"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type (
	// Resolver handles dependency injection and service initialization
	Resolver struct {
		config *workerConfig.Config
		db     *gorm.DB
	}
)

// NewResolver creates a new resolver instance
func NewResolver(cfg *workerConfig.Config) *Resolver {
	return &Resolver{
		config: cfg,
	}
}

// NewResolverWithDB creates a new resolver instance with database connection
func NewResolverWithDB(cfg *workerConfig.Config, db *gorm.DB) *Resolver {
	return &Resolver{
		config: cfg,
		db:     db,
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

// ResolveFileService creates a file service instance
func (r *Resolver) ResolveFileService(ctx context.Context) (service.FileService, error) {
	log.Info("Resolving file service")

	// Resolve all repository dependencies
	contractorRepo, err := r.ResolveContractorRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve contractor repository: %w", err)
	}

	contractorProjectRepo, err := r.ResolveContractorProjectRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve contractor project repository: %w", err)
	}

	projectRepo, err := r.ResolveProjectRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project repository: %w", err)
	}

	siteRepo, err := r.ResolveSiteRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve site repository: %w", err)
	}

	documentGroupRepo, err := r.ResolveDocumentGroupRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve document group repository: %w", err)
	}

	documentRepo, err := r.ResolveDocumentRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve document repository: %w", err)
	}

	fileRepo, err := r.ResolveFileRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file repository: %w", err)
	}

	// Create and return file service with all dependencies
	fileService := service.NewFileService(contractorRepo, contractorProjectRepo, projectRepo, siteRepo, documentGroupRepo, documentRepo, fileRepo)
	log.Info("File service resolved successfully")

	return fileService, nil
}

// ResolveS3Service creates an S3 service instance
func (r *Resolver) ResolveS3Service(ctx context.Context) (service.S3Service, error) {
	log.Info("Resolving S3 service")

	// Create S3 client
	s3Client, err := r.ResolveS3Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve S3 client: %w", err)
	}

	// Create AWS config for multi-region support
	var awsConfig aws.Config
	if r.config.AWSAccessKeyID != "" && r.config.AWSSecretAccessKey != "" {
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(r.config.AWSRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				r.config.AWSAccessKeyID,
				r.config.AWSSecretAccessKey,
				"", // session token
			)),
		)
	} else {
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(r.config.AWSRegion),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Resolve file service dependency
	fileService, err := r.ResolveFileService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file service: %w", err)
	}

	// Create and return S3 service with multi-region support
	s3Service := service.NewS3Service(s3Client, awsConfig, r.config.AWSAccessKeyID, r.config.AWSSecretAccessKey, fileService)
	log.Info("S3 service resolved successfully with multi-region support")

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

	// Create contractor repository
	contractorRepo, err := r.ResolveContractorRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve contractor repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create user_contractor repository
	userContractorRepo, err := r.ResolveUserContractorRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve user contractor repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create viewer_contractor repository
	viewerContractorRepo, err := r.ResolveViewerContractorRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve viewer contractor repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create contractor_project repository
	contractorProjectRepo, err := r.ResolveContractorProjectRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve contractor project repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create project repository
	projectRepo, err := r.ResolveProjectRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve project repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create site repository
	siteRepo, err := r.ResolveSiteRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve site repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create document group repository
	documentGroupRepo, err := r.ResolveDocumentGroupRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve document group repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create document repository
	documentRepo, err := r.ResolveDocumentRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve document repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create file repository
	fileRepo, err := r.ResolveFileRepository(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to resolve file repository, using null cleansing service")
		return service.NewNullCleansingService()
	}

	// Create and return cleansing service with all dependencies
	cleansingService := service.NewCleansingService(
		s3Service,
		contractorRepo,
		userContractorRepo,
		viewerContractorRepo,
		contractorProjectRepo,
		projectRepo,
		siteRepo,
		documentGroupRepo,
		documentRepo,
		fileRepo,
	)
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

// ResolveDatabase creates and returns a database connection
func (r *Resolver) ResolveDatabase(ctx context.Context) (*gorm.DB, error) {
	if r.db != nil {
		return r.db, nil
	}

	log.Info("Initializing database connection")
	db, err := database.NewConnection(r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	r.db = db
	return db, nil
}

// ResolveContractorRepository creates and returns a contractor repository
func (r *Resolver) ResolveContractorRepository(ctx context.Context) (repository.ContractorRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewContractorRepository(db), nil
}

// ResolveProjectRepository creates and returns a project repository
func (r *Resolver) ResolveProjectRepository(ctx context.Context) (repository.ProjectRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewProjectRepository(db), nil
}

// ResolveSiteRepository creates and returns a site repository
func (r *Resolver) ResolveSiteRepository(ctx context.Context) (repository.SiteRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewSiteRepository(db), nil
}

// ResolveDocumentGroupRepository creates and returns a document group repository
func (r *Resolver) ResolveDocumentGroupRepository(ctx context.Context) (repository.DocumentGroupRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewDocumentGroupRepository(db), nil
}

// ResolveDocumentRepository creates and returns a document repository
func (r *Resolver) ResolveDocumentRepository(ctx context.Context) (repository.DocumentRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewDocumentRepository(db), nil
}

// ResolveFileRepository creates and returns a file repository
func (r *Resolver) ResolveFileRepository(ctx context.Context) (repository.FileRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewFileRepository(db), nil
}

// ResolveContractorProjectRepository creates and returns a contractor project repository
func (r *Resolver) ResolveContractorProjectRepository(ctx context.Context) (repository.ContractorProjectRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewContractorProjectRepository(db), nil
}

// ResolveUserContractorRepository creates and returns a user_contractor repository
func (r *Resolver) ResolveUserContractorRepository(ctx context.Context) (repository.UserContractorRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewUserContractorRepository(db), nil
}

// ResolveViewerContractorRepository creates and returns a viewer_contractor repository
func (r *Resolver) ResolveViewerContractorRepository(ctx context.Context) (repository.ViewerContractorRepository, error) {
	db, err := r.ResolveDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return repository.NewViewerContractorRepository(db), nil
}
