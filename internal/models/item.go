package models

import "github.com/shopspring/decimal"

// Itemized item on a receipt
//
// swagger:model
type Item struct {
	BaseModel

	// Amount the item costs
	//
	// required: true
	Amount decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`

	ChargedToUser User `json:"-"`

	// User foreign key
	//
	// required: true
	ChargedToUserId uint `json:"chargedToUserId" gorm:"not null"`

	// Is taxed (not used)
	//
	// required: false
	IsTaxed bool `gorm:"not null"`

	// Item name
	//
	// required: true
	Name string `json:"name" gorm:"not null"`

	Receipt Receipt `json:"-"`

	// Receipt foreign key
	//
	// required: true
	ReceiptId uint `json:"receiptId"`

	// Receipt status
	//
	// required: true
	Status ItemStatus `gorm:"default:'OPEN'; not null" json:"status"`
}
