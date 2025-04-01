package models

type CustomField struct {
	BaseModel
	Name        string              `gorm:"not null" json:"name"`
	Type        CustomFieldType     `gorm:"not null" json:"type"`
	Description string              `json:"description"`
	Options     []CustomFieldOption `json:"options"`
}
