package models

type Item struct {
	BaseModel
	Name            string `gorm:"not null"`
	ChargedToUser   User
	ChargedToUserId uint
	ReceiptId       uint `gorm:"not null"`
	Receipt         Receipt
	Amount          uint `gorm:"not null"`
	IsTaxed         bool `gorm:"not null"`
}
