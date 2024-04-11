package models

type Comment struct {
	BaseModel
	Comment        string    `gorm:"type:varchar(500); not null" json:"comment"`
	Receipt        Receipt   `json:"-"`
	ReceiptId      uint      `json:"receiptId"`
	User           User      `json:"-"`
	UserId         *uint     `json:"userId"`
	AdditionalInfo string    `gorm:"type:varchar(500)" json:"additionalInfo"`
	CommentId      *uint     `json:"commentId"`
	Replies        []Comment `json:"replies"`
}
