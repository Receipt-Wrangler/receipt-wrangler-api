package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Receipt struct {
	BaseModel
	Name         string          `gorm:"not null" json:"name"`
	Amount       decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`
	Date         time.Time       `gorm:"not null" json:"date"`
	ImgPath      string          `json:"-"`
	PaidByUserID uint            `json:"paidByUserId"`
	PaidByUser   User            `json:"-"`
	IsResolved   bool            `gorm:"default: false" json:"isResolved"`
	GroupId      uint            `gorm:"not null" json:"groupId"`
	Group        Group           `json:"-"`
	Tags         []Tag           `gorm:"many2many:receipt_tags" json:"tags"`
	Categories   []Category      `gorm:"many2many:receipt_categories" json:"categories"`
	ImageFiles   []FileData      `json:"imageFiles"`
	ReceiptItems []Item          `json:"receiptItems"`
}

func (r *Receipt) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Where("receipt_id IS NULL").Delete(&Item{})
	return
}
