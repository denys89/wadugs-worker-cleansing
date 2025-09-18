package entity

type (
	DocumentGroups []DocumentGroup

	DocumentGroup struct {
		Id               int64  `json:"id" gorm:"column:id"`
		SiteId           int64  `json:"site_id" gorm:"column:site_id"`
		Name             string `json:"name" gorm:"column:name"`
		ProcessedName    string `json:"processed_name" gorm:"column:processed_name"`
		ProcessedAt      int64  `json:"processed_at" gorm:"column:processed_at"`
		ProcessedErrCode string `json:"processed_err_code" gorm:"column:processed_err_code"`
		ProcessedRemarks string `json:"processed_remarks" gorm:"column:processed_remarks"`
		Category         string `json:"category" gorm:"column:category"`
		Status           int8   `json:"status" gorm:"column:status"`
		Showed           int8   `json:"showed" gorm:"column:showed"`
		Progress         int8   `json:"progress" gorm:"column:progress"`
		MetaData
	}

	DocumentGroupV2 struct {
		Id               int64  `json:"id" gorm:"column:id"`
		SiteId           int64  `json:"site_id" gorm:"column:site_id"`
		Name             string `json:"name" gorm:"column:name"`
		ProcessedName    string `json:"processed_name" gorm:"column:processed_name"`
		ProcessedAt      int64  `json:"processed_at" gorm:"column:processed_at"`
		ProcessedErrCode string `json:"processed_err_code" gorm:"column:processed_err_code"`
		ProcessedRemarks string `json:"processed_remarks" gorm:"column:processed_remarks"`
		Category         string `json:"category" gorm:"column:category"`
		Status           int8   `json:"status" gorm:"column:status"`
		Showed           int8   `json:"showed" gorm:"column:showed"`
		Progress         int8   `json:"progress" gorm:"column:progress"`
		SiteName         string `json:"site_name" gorm:"column:site_name"`
		SiteAlias        string `json:"site_alias" gorm:"column:site_alias"`
	}
)

func (d DocumentGroup) TableName() string {
	return "document_group"
}

func (d DocumentGroup) PrimaryKey() string {
	return "id"
}

func (d DocumentGroup) GetAllowedOrderFields() []string {
	return []string{"id", "site_id", "name", "processed_name", "category", "status", "progress", "created_at", "updated_at"}
}

func (d DocumentGroup) GetAllowedWhereFields() []string {
	return []string{"id", "site_id", "name", "processed_name", "category", "status", "progress", "processed_err_code"}
}