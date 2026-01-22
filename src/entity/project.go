package entity

const (
	DefaultGCRSName = "Custom"
	DefaultPCRSName = "Custom"
	DefaultGCRSID   = int64(0)
	DefaultPCRSID   = int64(0)
)

type (
	Projects []Project

	Project struct {
		Id               int64  `json:"id" gorm:"column:id;primaryKey"`
		ContractorId     int64  `json:"contractor_id" gorm:"column:contractor_id"`
		Name             string `json:"name" gorm:"column:name"`
		Code             string `json:"code" gorm:"column:code"`
		ProjectNumber    string `json:"project_number" gorm:"column:project_number"`
		Alias            string `json:"alias" gorm:"column:alias"`
		Status           int8   `json:"status" gorm:"column:status"` // 0: NOT ACTIVE, 1: ACTIVE
		IsDeleted        bool   `json:"is_deleted" gorm:"column:is_deleted;default:false"`
		GCRS             string `json:"g_crs" gorm:"column:g_crs"`
		PCRS             string `json:"p_crs" gorm:"column:p_crs"`
		Coordinates      *int8  `json:"coordinates" gorm:"column:coordinates"` // 0:longitude-latitude 1: Easting-Nothing 2: Both
		Block            string `json:"block" gorm:"column:block"`
		Start            int64  `json:"start" gorm:"column:start"`
		End              int64  `json:"end" gorm:"column:end"`
		VesselNumber     int8   `json:"vessel_number" gorm:"column:vessel_number"`
		ClientUserNumber int8   `json:"client_user_number" gorm:"column:client_user_number"`
		UploaderNumber   int8   `json:"uploader_number" gorm:"column:uploader_number"`
		FileSizeCapacity int64  `json:"file_size_capacity" gorm:"column:file_size_capacity"`
		FileSizeUsage    int64  `json:"file_size_usage" gorm:"column:file_size_usage"`
		MetaData
	}

	CRS struct {
		Id     int64  `json:"id"`
		Name   string `json:"name"`
		Syntax string `json:"syntax"`
	}
)

func (p Project) TableName() string {
	return "project"
}

func (p Project) PrimaryKey() string {
	return "id"
}

func (p Project) GetAllowedOrderFields() []string {
	return []string{"id", "name", "code", "status", "block", "start", "end", "created_at", "updated_at"}
}

func (p Project) GetAllowedWhereFields() []string {
	return []string{"id", "name", "code", "status", "is_deleted", "block", "start", "end", "created_at", "updated_at", "created_by", "updated_by", "contractor_id"}
}
