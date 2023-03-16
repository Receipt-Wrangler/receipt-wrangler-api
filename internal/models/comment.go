package models

type Comment struct {
	BaseModel
	Comment        string  `gorm:"type:varchar(500)" json:"comment"`
	Receipt        Receipt `json:"-"`
	ReceiptId      uint    `json:"receiptId"`
	User           User    `json:"-"`
	UserId         *uint   `json:"userId"`
	AdditionalInfo string  `gorm:"type:varchar(500)" json:"additionalInfo"`
}
