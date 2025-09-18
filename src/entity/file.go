package entity

type (
	Files   []File
	FilesV2 []FileV2

	File struct {
		Id         int64  `json:"id" gorm:"column:id;primaryKey"`
		DocumentId int64  `json:"document_id" gorm:"column:document_id"`
		Size       int64  `json:"size" gorm:"column:size"`
		RasterSize int64  `json:"raster_size" gorm:"column:raster_size"`
		Checksum   string `json:"checksum" gorm:"column:checksum"` // Content file checksum. Only allow re-upload when checksum changes
		Name       string `json:"name" gorm:"column:name"`         // filename.ini
		Status     int8   `json:"status" gorm:"column:status"`     // '10: ACTIVE, 20: UPLOAD QUOTA OVER LIMIT',
		MetaData
	}

	FileV2 struct {
		Id           int64  `json:"id" gorm:"column:id;primaryKey"`
		DocumentId   int64  `json:"document_id" gorm:"column:document_id"`
		Size         int64  `json:"size" gorm:"column:size"`
		RasterSize   int64  `json:"raster_size" gorm:"column:raster_size"`
		Checksum     string `json:"checksum" gorm:"column:checksum"` // Content file checksum. Only allow re-upload when checksum changes
		Name         string `json:"name" gorm:"column:name"`         // filename.ini
		Status       int8   `json:"status" gorm:"column:status"`
		Category     string `json:"category" gorm:"column:category"`
		Sensor       int    `json:"sensor" gorm:"column:sensor"`
		DocumentName string `json:"document_name" gorm:"column:document_name"`
		MetaData
	}

	UploaderContractor struct {
		Id           int64 `json:"id" gorm:"column:id;primaryKey"` // uploader id
		UploaderId   int64 `json:"uploader_id" gorm:"column:uploader_id"`
		ContractorId int64 `json:"contractor_id" gorm:"column:contractor_id"`
		ProjectId    int64 `json:"project_id" gorm:"column:project_id"`
	}

	UploaderContractorUsage struct {
		Id               int64 `json:"id" gorm:"column:id;primaryKey"` // uploader id
		UploaderId       int64 `json:"uploader_id" gorm:"column:uploader_id"`
		ProjectId        int64 `json:"project_id" gorm:"column:project_id"`
		FileSizeUsage    int64 `json:"file_size_usage" gorm:"column:file_size_usage"`
		FileSizeCapacity int64 `json:"file_size_capacity" gorm:"column:file_size_capacity"`
	}
)

func (f File) TableName() string {
	return "file"
}

func (f File) PrimaryKey() string {
	return "id"
}

func (f File) GetAllowedOrderFields() []string {
	return []string{"id", "document_id", "key", "status", "created_at", "updated_at"}
}

func (f File) GetAllowedWhereFields() []string {
	return []string{"id", "document_id", "key", "status", "created_at", "updated_at", "created_by", "updated_by"}
}

func (uk FileV2) TableName() string {
	return File{}.TableName()
}

func (uk FileV2) PrimaryKey() string {
	return File{}.PrimaryKey()
}

func (uk FileV2) GetAllowedOrderFields() []string {
	return File{}.GetAllowedOrderFields()
}

func (uk FileV2) GetAllowedWhereFields() []string {
	return File{}.GetAllowedWhereFields()
}