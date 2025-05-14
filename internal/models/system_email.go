package models

type SystemEmail struct {
	BaseModel
	Host        string `gorm:"not null;" json:"host"`
	Port        string `gorm:"not null;" json:"port"`
	Username    string `gorm:"not null;" json:"username"`
	Password    string `gorm:"not null;" json:"-"`
	UseStartTLS bool   `gorm:"not null; default: false;" json:"useStartTLS"`
}

type SystemEmailWithPassword struct {
	BaseModel
	SystemEmail
	Password string `json:"password"`
}
