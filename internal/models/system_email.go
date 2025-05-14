package models

type SystemEmail struct {
	BaseModel
	Host        string `gorm:"not null;" json:"host"`
	Port        string `gorm:"not null;" json:"port"`
	Username    string `gorm:"not null;" json:"username"`
	Password    string `gorm:"not null;" json:"-"`
	UseStartTls bool   `gorm:"not null; default: false;" json:"useStartTls"`
}

type SystemEmailWithPassword struct {
	BaseModel
	SystemEmail
	Password string `json:"password"`
}
