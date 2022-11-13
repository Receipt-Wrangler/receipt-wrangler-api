package models

import "github.com/shopspring/decimal"

type Item struct {
	BaseModel
	Name            string          `json:"name" gorm:"not null"`
	ChargedToUser   User            `json:"-"`
	ChargedToUserId uint            `json:"chargedToUserId" gorm:"not null"`
	ReceiptId       uint            `json:"receiptId"`
	Receipt         Receipt         `json:"-"`
	Amount          decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`
	IsTaxed         bool            `gorm:"not null"`
}
