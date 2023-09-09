package models

type Email struct {
	Email string `gorm:"not null;" json:"email"`
}
