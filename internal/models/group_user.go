package models

type GroupUser struct {
	User    User
	UserId  uint `gorm:"primaryKey"`
	Group   Group
	GroupId uint `gorm:"primaryKey"`
}
