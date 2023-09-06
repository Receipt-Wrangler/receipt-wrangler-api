package models

type GroupSettings struct {
	BaseModel
	GroupId uint `gorm:"not null;unique" json:"groupId"`
}
