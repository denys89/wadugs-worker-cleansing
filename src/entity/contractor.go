package entity

type (
	Contractors []Contractor

	Contractor struct {
		Id                    int64  `json:"id" gorm:"column:id;primaryKey"`
		Name                  string `json:"name" gorm:"column:name"`
		Alias                 string `json:"alias" gorm:"column:alias"`
		Status                int8   `json:"status" gorm:"column:status"` // 0: NOT ACTIVE, 1: ACTIVE
		CountryCode           string `json:"country_code" gorm:"column:country_code"`
		DBHost                string `json:"db_host" gorm:"column:db_host"`
		DBName                string `json:"db_name" gorm:"column:db_name"`
		DBUser                string `json:"db_user" gorm:"column:db_user"`
		DBPass                string `json:"db_pass" gorm:"column:db_pass"`
		Logo                  string `json:"logo" gorm:"column:logo"`
		AwsIamAccessKeyId     string `json:"aws_iam_access_key_id" gorm:"column:aws_iam_access_key_id"`
		AwsIamSecretAccessKey string `json:"aws_iam_secret_access_key" gorm:"column:aws_iam_secret_access_key"`
		AwsBucketName         string `json:"aws_bucket_name" gorm:"column:aws_bucket_name"`
		AwsBucketRegion       string `json:"aws_bucket_region" gorm:"column:aws_bucket_region"`
		LambdaUrl             string `json:"lambda_url" gorm:"column:lambda_url"`
		LambdaLog             string `json:"lambda_log" gorm:"column:lambda_log"`
		ViewerNumber          uint8  `json:"viewer_number" gorm:"column:viewer_number"`
		MetaData
	}
)

func (c Contractor) TableName() string {
	return "contractor"
}

func (c Contractor) PrimaryKey() string {
	return "id"
}

func (c Contractor) GetAllowedOrderFields() []string {
	return []string{"id", "name", "status", "country_code", "created_at", "updated_at"}
}

func (c Contractor) GetAllowedWhereFields() []string {
	return []string{"id", "name", "status", "country_code", "created_at", "updated_at", "created_by", "updated_by"}
}