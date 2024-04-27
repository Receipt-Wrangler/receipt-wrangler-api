package models

type SystemEmail struct {
	BaseModel
	Host     string `gorm:"not null;" json:"host"`
	Port     string `gorm:"not null;" json:"port"`
	Username string `gorm:"not null;" json:"username"`
	Password string `gorm:"not null;" json:"-"`
}
