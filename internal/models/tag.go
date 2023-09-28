package models

type Tag struct {
	BaseModel
	Name        string `gorm:"not null; uniqueIndex" json:"name"`
	Description string `json:"description"`
}
