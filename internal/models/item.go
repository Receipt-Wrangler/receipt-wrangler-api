package models

import "github.com/shopspring/decimal"

type Item struct {
	BaseModel
	Amount          decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`
	ChargedToUser   User            `json:"-"`
	ChargedToUserId uint            `json:"chargedToUserId" gorm:"not null"`
	IsTaxed         bool            `gorm:"not null"`
	Name            string          `json:"name" gorm:"not null"`
	Receipt         Receipt         `json:"-"`
	ReceiptId       uint            `json:"receiptId"`
	Status          ItemStatus      `gorm:"default:'OPEN'; not null" json:"status"`
}
