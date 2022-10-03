package models

import "time"

type Receipt struct {
	BaseModel
	Name          string    `gorm:"not null" json:"name"`
	Amount        float64   `gorm:"not null" json:"amount"`
	Date          time.Time `gorm:"not null" json:"date"`
	ImgPath       string    `json:"-"`
	PaidByUserID  uint      `json:"paidByUserId"`
	PaidByUser    User      `json:"-"`
	OwnedByUserID uint
	OwnedByUser   User  `json:"-"`
	Tags          []Tag `gorm:"many2many:receipt_tags" json:"tags"`
}
