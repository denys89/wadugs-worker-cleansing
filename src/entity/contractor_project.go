package entity

type (
	ContractorProjects []ContractorProject

	ContractorProject struct {
		Id           int64 `json:"id" gorm:"column:id;primaryKey"`
		ContractorId int64 `json:"contractor_id" gorm:"column:contractor_id"`
		ProjectId    int64 `json:"project_id" gorm:"column:project_id"`
		MetaData
	}
)

func (c ContractorProject) TableName() string {
	return "contractor_project"
}

func (c ContractorProject) PrimaryKey() string {
	return "id"
}
