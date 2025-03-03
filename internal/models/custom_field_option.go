package models

type CustomFieldOption struct {
	BaseModel
	Value         string      `gorm:"not null" json:"value"`
	CustomField   CustomField `json:"-"`
	CustomFieldId uint        `gorm:"not null" json:"customFieldId"`
}
