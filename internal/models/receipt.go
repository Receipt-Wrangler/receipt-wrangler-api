package models

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type Receipt struct {
	BaseModel
	Name         string             `gorm:"not null" json:"name"`
	Amount       decimal.Decimal    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Date         time.Time          `gorm:"not null" json:"date"`
	ResolvedDate *time.Time         `json:"resolvedDate"`
	PaidByUserID uint               `json:"paidByUserId"`
	PaidByUser   User               `json:"-"`
	Status       ReceiptStatus      `gorm:"default:'OPEN';not null" json:"status"`
	GroupId      uint               `gorm:"not null" json:"groupId"`
	Group        Group              `json:"-"`
	Categories   []Category         `gorm:"many2many:receipt_categories" json:"categories"`
	Tags         []Tag              `gorm:"many2many:receipt_tags" json:"tags"`
	ImageFiles   []FileData         `json:"imageFiles"`
	ReceiptItems []Item             `json:"receiptItems"`
	Comments     []Comment          `json:"comments"`
	CustomFields []CustomFieldValue `json:"customFields"`
}

func (r *Receipt) ToString() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
