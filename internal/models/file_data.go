package models

type FileData struct {
	BaseModel
	Name      string  `json:"name"`
	FileType  string  `json:"fileType"`
	Size      uint    `json:"size"`
	ReceiptId uint    `json:"receiptId"`
	Receipt   Receipt `json:"-"`
}
