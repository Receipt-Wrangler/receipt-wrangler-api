package commands

type UpsertCustomFieldValueCommand struct {
	ReceiptId     uint   `json:"receiptId"`
	CustomFieldId uint   `json:"customFieldId"`
	Value         string `json:"value"`
}
