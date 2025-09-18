package entity

type (
	Documents   []Document
	DocumentsV2 []DocumentV2

	DocumentV2 struct {
		Id        int64  `json:"id" gorm:"column:id;primaryKey"`
		SiteId    int64  `json:"site_id" gorm:"column:site_id"`
		Name      string `json:"name" gorm:"column:name"`
		IsChecked bool   `json:"is_checked" gorm:"column:is_checked"`
		Status    int8   `json:"status" gorm:"column:status"`     // 0: NOT ACTIVE, 1: ACTIVE
		Progress  int8   `json:"progress" gorm:"column:progress"` // 1=Belum Diproses, 2=Sedang Diproses, 3=Selesai Diproses
		Key       string `json:"key" gorm:"column:key"`           // AWS S3 object key : /namesite_01/processed/Binning/filename
		Category  string `json:"category" gorm:"column:category"` // "Raster (Depth)", "Raster (Other)", "Side Scan Sonar", "Sub-Bottom Profiler", "Multi-Channel Seismic", "Magnetometer", "SIngle Beam Echo Sounder", "Interpretation", "Image", "Soil Sample", "Line Route", "Boundary",  "Vector"
		SiteCode  string `json:"site_code" gorm:"column:site_code"`
		SiteName  string `json:"site_name" gorm:"column:site_name"`
		MetaData
	}

	Document struct {
		Id         int64  `json:"id" gorm:"primaryKey;autoIncrement"`
		GroupID    int64  `json:"group_id" gorm:"not null"`
		Name       string `json:"name" gorm:"size:255;not null;comment:filename"`
		IsChecked  bool   `json:"is_checked" gorm:"column:is_checked"`
		Status     int8   `json:"status" gorm:"column:status"` // 1: PROCESSED, 0: NOT PROCESSED
		Attachment string `json:"attachment" gorm:"size:255;not null;comment:filename"`
		MetaData
	}

	DocumentV3 struct {
		Id         int64  `json:"id" gorm:"primaryKey;autoIncrement"`
		GroupID    int64  `json:"group_id" gorm:"not null"`
		Name       string `json:"name" gorm:"size:255;not null;comment:filename"`
		Attachment string `json:"attachment" gorm:"size:255;not null;comment:filename"`
		IsChecked  bool   `json:"is_checked" gorm:"column:is_checked"`
		Status     int8   `json:"status" gorm:"column:status"` // 1: PROCESSED, 0: NOT PROCESSED
		Progress   int8   `json:"progress" gorm:"column:progress"`
		MetaData
	}

	DocumentProcess struct {
		Id                 int64  `json:"id" gorm:"column:id"`             // document id
		Name               string `json:"name" gorm:"column:name"`         // document name
		GroupId            int64  `json:"group_id" gorm:"column:group_id"` // document group id
		Category           string `json:"category" gorm:"column:category"` // document group category
		Progress           int8   `json:"progress" gorm:"column:progress"`
		SiteId             int64  `json:"site_id" gorm:"column:site_id"`
		SiteCode           string `json:"site_code" gorm:"column:site_code"`
		SiteAlias          string `json:"site_alias" gorm:"column:site_alias"`
		SiteName           string `json:"site_name" gorm:"column:site_name"`
		ProjectId          int64  `json:"project_id" gorm:"column:project_id"`
		ProjectCode        string `json:"project_code" gorm:"column:project_code"`
		ProjectAlias       string `json:"project_alias" gorm:"column:project_alias"`
		ProjectName        string `json:"project_name" gorm:"column:project_name"`
		ContractorId       int64  `json:"contractor_id" gorm:"column:contractor_id"`
		UploaderId         int64  `json:"uploader_id" gorm:"column:uploader_id"`
		ContractorAlias    string `json:"contractor_alias" gorm:"column:contractor_alias"`
		AwsAccessKeyId     string `json:"aws_iam_access_key_id" gorm:"column:aws_iam_access_key_id"`
		AwsSecretAccessKey string `json:"aws_iam_secret_access_key" gorm:"column:aws_iam_secret_access_key"`
		AwsBucketName      string `json:"aws_bucket_name" gorm:"column:aws_bucket_name"`
		AwsBucketRegion    string `json:"aws_bucket_region" gorm:"column:aws_bucket_region"`
		DBName             string `json:"db_name" gorm:"column:db_name"`
		DBUser             string `json:"db_user" gorm:"column:db_user"`
		DBPassword         string `json:"db_pass" gorm:"column:db_pass"`
		DBHost             string `json:"db_host" gorm:"column:db_host"`
		Gcrs               string `json:"gcrs" gorm:"column:gcrs"`
		Pcrs               string `json:"pcrs" gorm:"column:pcrs"`
		TotalSize          int64  `json:"total_size" gorm:"column:total_size"`
	}

	DocumentUpload struct {
		S3BucketName    string `gorm:"column:aws_bucket_name"`
		AWSBucketRegion string `gorm:"column:aws_bucket_region"`
		SiteCode        string `gorm:"column:site_code"`
		ProjectCode     string `gorm:"column:project_code"`
		ProjectId       int64  `gorm:"column:project_id"`
	}

	DocumentGroupUpload struct {
		S3BucketName               string `gorm:"column:aws_bucket_name"`
		AWSBucketRegion            string `gorm:"column:aws_bucket_region"`
		SiteCode                   string `gorm:"column:site_code"`
		ProjectCode                string `gorm:"column:project_code"`
		ProjectId                  int64  `gorm:"column:project_id"`
		DocumentGroupName          string `gorm:"column:document_group_name"`
		DocumentGroupProcessedName string `gorm:"column:document_group_processed_name"`
		DocumentGroupCategory      string `gorm:"column:document_group_category"`
		DocumentGroupProgress      int64  `gorm:"column:document_group_progress"`
	}

	DocumentFile struct {
		GroupId    int64  `gorm:"column:group_id"`
		DocumentId int64  `gorm:"column:document_id"`
		FileName   string `gorm:"column:name"`
	}
)

func (d Document) TableName() string {
	return "document"
}

func (d Document) PrimaryKey() string {
	return "id"
}

func (d Document) GetAllowedOrderFields() []string {
	return []string{"id", "group_id", "name", "created_at", "updated_at"}
}

func (d Document) GetAllowedWhereFields() []string {
	return []string{"id", "name", "group_id", "is_checked"}
}

func (uk DocumentV2) TableName() string {
	return Document{}.TableName()
}

func (uk DocumentV2) PrimaryKey() string {
	return Document{}.PrimaryKey()
}

func (uk DocumentV2) GetAllowedOrderFields() []string {
	return Document{}.GetAllowedOrderFields()
}

func (uk DocumentV2) GetAllowedWhereFields() []string {
	return Document{}.GetAllowedWhereFields()
}