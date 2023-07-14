package models

// File data for images on a receipt
//
// swagger:model
type FileData struct {
	BaseModel

	// File name
	//
	// required: false
	Name string `json:"name"`

	// Image data
	//
	// required: true
	ImageData []byte `gorm:"-" json:"imageData"`

	// MIME file type
	//
	// required: false
	FileType string `json:"fileType"`

	// File size
	//
	// required: false
	Size uint `json:"size"`

	// Receipt foreign key
	//
	// required: true
	ReceiptId uint    `json:"receiptId"`
	Receipt   Receipt `json:"-"`
}
