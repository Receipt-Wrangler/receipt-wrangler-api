package models

import "github.com/shopspring/decimal"

type Item struct {
	BaseModel
	Name            string `json:"name" gorm:"not null"`
	ChargedToUser   User   `json:"chargedToUser"`
	ChargedToUserId uint   `json:"chargedToUserId" gorm:"not null"`
	ReceiptId       uint   `json:"receiptId" gorm:"not null"`
	Receipt         Receipt
	Amount          decimal.Decimal `json:"amount" sql:"type:decimal(20,3);"`
	IsTaxed         bool            `gorm:"not null"`
}
