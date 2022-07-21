package models

type Group struct {
	BaseModel
	Name string `gorm:"not null" json:"name"`
}
