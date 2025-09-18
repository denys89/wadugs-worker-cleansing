package entity

type MetaData struct {
	CreatedAt int64 `json:"created_at" gorm:"column:created_at"`
	UpdatedAt int64 `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy int64 `json:"created_by" gorm:"column:created_by"`
	UpdatedBy int64 `json:"updated_by" gorm:"column:updated_by"`
}