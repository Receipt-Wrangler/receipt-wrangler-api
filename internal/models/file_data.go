package models

type FileData struct {
	BaseModel
	Name      string
	ImageData []byte `gorm:"-"`
	FileType  string
	Size      uint
	ReceiptId uint
	Receipt   Receipt
}
