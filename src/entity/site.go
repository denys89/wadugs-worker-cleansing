package entity

type (
	Sites   []Site
	SitesV2 []SiteV2

	Site struct {
		Id        int64  `json:"id" gorm:"column:id;primaryKey"`
		Code      string `json:"code" gorm:"column:code"`
		Alias     string `json:"alias" gorm:"column:alias"`
		Name      string `json:"name" gorm:"column:name"`
		Status    int8   `json:"status" gorm:"column:status"` // 0: NOT ACTIVE, 1: ACTIVE
		Area      string `json:"area" gorm:"column:area"`
		ProjectId int64  `json:"project_id" gorm:"column:project_id"`
		MetaData
	}

	SiteV2 struct {
		Id              int64  `json:"id" gorm:"column:id;primaryKey"`
		Code            string `json:"code" gorm:"column:code"`
		Name            string `json:"name" gorm:"column:name"`
		Status          int8   `json:"status" gorm:"column:status"` // 0: NOT ACTIVE, 1: ACTIVE
		Area            string `json:"area" gorm:"column:area"`
		ProjectId       int64  `json:"project_id" gorm:"column:project_id"`
		ProjectName     string `json:"project_name" gorm:"column:project_name"`
		ContractorName  string `json:"contractor_name" gorm:"column:contractor_name"`
		GCRS            string `json:"g_crs" gorm:"column:g_crs"`
		PCRS            string `json:"p_crs" gorm:"column:p_crs"`
		Coordinates     int    `json:"coordinates" gorm:"column:coordinates"`
		ContractorAlias string `json:"contractor_alias" gorm:"column:contractor_alias"`
		ProjectAlias    string `json:"project_alias" gorm:"column:project_alias"`
		SiteAlias       string `json:"site_alias" gorm:"column:alias"`
		LambdaUrl       string `json:"lambda_url" gorm:"column:lambda_url"`
		MetaData
	}

	SiteProject struct {
		Id         int64 `json:"id" gorm:"column:id;primaryKey"` // uploader_id
		SiteId     int64 `json:"site_id" gorm:"column:site_id"`
		ProjectId  int64 `json:"project_id" gorm:"column:project_id"`
		UploaderId int64 `json:"uploader_id" gorm:"column:uploader_id"`
	}
)

func (s Site) TableName() string {
	return "site"
}

func (s Site) PrimaryKey() string {
	return "id"
}

func (s Site) GetAllowedOrderFields() []string {
	return []string{"id", "code", "name", "status", "area", "project_id", "created_at", "updated_at"}
}

func (s Site) GetAllowedWhereFields() []string {
	return []string{"id", "code", "name", "status", "area", "project_id", "created_at", "updated_at", "created_by", "updated_by"}
}

func (uk SiteV2) TableName() string {
	return Site{}.TableName()
}

func (uk SiteV2) PrimaryKey() string {
	return Site{}.PrimaryKey()
}

func (uk SiteV2) GetAllowedOrderFields() []string {
	return Site{}.GetAllowedOrderFields()
}

func (uk SiteV2) GetAllowedWhereFields() []string {
	return Site{}.GetAllowedWhereFields()
}