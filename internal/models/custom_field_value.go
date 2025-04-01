package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type CustomFieldValue struct {
	BaseModel
	Receipt       Receipt          `json:"-"`
	ReceiptId     uint             `json:"receiptId"`
	CustomField   CustomField      `json:"-"`
	CustomFieldId uint             `json:"customFieldId"`
	StringValue   *string          `json:"stringValue"`
	DateValue     *time.Time       `json:"dateValue"`
	SelectValue   *uint            `json:"selectValue"`
	CurrencyValue *decimal.Decimal `json:"currencyValue"`
	BooleanValue  *bool            `json:"booleanValue"`
}
