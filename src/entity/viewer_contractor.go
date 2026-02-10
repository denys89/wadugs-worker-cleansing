package entity

type (
	ViewerContractors []ViewerContractor

	ViewerContractor struct {
		Id           int64 `json:"id" gorm:"column:id;primaryKey"`
		ViewerId     int64 `json:"viewer_id" gorm:"column:viewer_id"`
		ContractorId int64 `json:"contractor_id" gorm:"column:contractor_id"`
		MetaData
	}
)

func (vc ViewerContractor) TableName() string {
	return "viewer_contractor"
}
