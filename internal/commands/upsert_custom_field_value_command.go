package commands

import (
	"github.com/shopspring/decimal"
	"time"
)

type UpsertCustomFieldValueCommand struct {
	ReceiptId     uint             `json:"receiptId"`
	CustomFieldId uint             `json:"customFieldId"`
	StringValue   *string          `json:"stringValue"`
	DateValue     *time.Time       `json:"dateValue"`
	SelectValue   *uint            `json:"selectValue"`
	CurrencyValue *decimal.Decimal `json:"currencyValue"`
	BooleanValue  *bool            `json:"booleanValue"`
}
