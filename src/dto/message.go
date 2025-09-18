package dto

const (
	// CleansingType constants for different deletion types
	CleansingTypeContractor = "contractor"
	CleansingTypeProject    = "project"
	CleansingTypeSite       = "site"
)

type (
	// CleansingMessage represents the message payload for data cleansing operations
	CleansingMessage struct {
		Type string `json:"type"` // contractor, project, or site
		ID   int64  `json:"id"`   // corresponding ID: contractor_id, project_id, or site_id
	}

	// CleansingResult represents the result of a cleansing operation
	CleansingResult struct {
		Type         string `json:"type"`
		ID           int64  `json:"id"`
		Success      bool   `json:"success"`
		Message      string `json:"message"`
		FilesDeleted int    `json:"files_deleted"`
		Error        string `json:"error,omitempty"`
	}

	// S3Object represents an S3 object to be deleted
	S3Object struct {
		Bucket string `json:"bucket"`
		Key    string `json:"key"`
	}

	// DeletionContext contains information needed for file deletion operations
	DeletionContext struct {
		Type        string     `json:"type"`
		ID          int64      `json:"id"`
		S3Objects   []S3Object `json:"s3_objects"`
		Description string     `json:"description"`
	}
)

// IsValidType checks if the cleansing type is valid
func (cm *CleansingMessage) IsValidType() bool {
	switch cm.Type {
	case CleansingTypeContractor, CleansingTypeProject, CleansingTypeSite:
		return true
	default:
		return false
	}
}

// GetDescription returns a human-readable description of the cleansing operation
func (cm *CleansingMessage) GetDescription() string {
	switch cm.Type {
	case CleansingTypeContractor:
		return "Deleting all files for contractor and its related projects and sites"
	case CleansingTypeProject:
		return "Deleting all files for project and its related sites"
	case CleansingTypeSite:
		return "Deleting all files for site"
	default:
		return "Unknown cleansing operation"
	}
}