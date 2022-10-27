package models

import "github.com/shopspring/decimal"

type Item struct {
	BaseModel
	Name            string `gorm:"not null"`
	ChargedToUser   User
	ChargedToUserId uint `gorm:"not null"`
	ReceiptId       uint `gorm:"not null"`
	Receipt         Receipt
	Amount          decimal.Decimal `json:"amount" sql:"type:decimal(20,3);"`
	IsTaxed         bool            `gorm:"not null"`
}
