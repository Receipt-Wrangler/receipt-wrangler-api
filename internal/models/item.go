package models

import "github.com/shopspring/decimal"

type Item struct {
	BaseModel
	Amount          decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`
	ChargedToUser   User            `json:"-"`
	ChargedToUserId *uint           `json:"chargedToUserId"`
	IsTaxed         bool            `gorm:"not null"`
	Name            string          `json:"name" gorm:"not null"`
	Receipt         Receipt         `json:"-"`
	ReceiptId       uint            `json:"receiptId"`
	Status          ItemStatus      `gorm:"default:'OPEN'; not null" json:"status"`
	Categories      []Category      `gorm:"many2many:item_categories" json:"categories"`
	Tags            []Tag           `gorm:"many2many:item_tags" json:"tags"`
	LinkedItems     []Item          `gorm:"many2many:item_linked_items" json:"linkedItems"`
}
