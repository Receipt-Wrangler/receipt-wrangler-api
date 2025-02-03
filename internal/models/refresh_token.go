package models

type RefreshToken struct {
	BaseModel
	UserId uint   `gorm:"not null`
	Token  string `gorm:"not null"`
	IsUsed bool   `gorm:"default:false"`
}
