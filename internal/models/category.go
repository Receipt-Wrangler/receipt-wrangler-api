package models

type Category struct {
	BaseModel
	Name string `gorm:"not null; uniqueIndex" json:"name"`
}
