package models

type Prompt struct {
	BaseModel
	Name        string `gorm:"not null; uniqueIndex" json:"name"`
	Description string `json:"description"`
	Prompt      string `gorm:"not null;" json:"prompt"`
}
