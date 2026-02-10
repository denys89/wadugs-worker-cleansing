package entity

type (
	UserContractors []UserContractor

	UserContractor struct {
		Id           int64 `json:"id" gorm:"column:id;primaryKey"`
		ContractorId int64 `json:"contractor_id" gorm:"column:contractor_id"`
		UserId       int64 `json:"user_id" gorm:"column:user_id"`
		IsDeleted    int8  `json:"is_deleted" gorm:"column:is_deleted"`
		MetaData
	}
)

func (uc UserContractor) TableName() string {
	return "user_contractor"
}
