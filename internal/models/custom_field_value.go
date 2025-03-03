package models

type CustomFieldValue struct {
	BaseModel
	Receipt       Receipt     `json:"-"`
	ReceiptId     uint        `json:"receiptId"`
	CustomField   CustomField `json:"-"`
	CustomFieldId uint        `json:"customFieldId"`
	Value         string      `json:"value"`
}
