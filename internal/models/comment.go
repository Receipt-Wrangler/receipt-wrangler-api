package models

// User comment left on receipts
//
// swagger:model
type Comment struct {
	BaseModel

	// Comment itself
	//
	// required: true
	Comment string `gorm:"type:varchar(500); not null" json:"comment"`

	Receipt Receipt `json:"-"`

	// Receipt foreign key
	//
	// required: true
	ReceiptId uint `json:"receiptId"`

	User User `json:"-"`

	// User foreign key
	//
	// required: true
	UserId *uint `json:"userId"`

	// Additional information about the comment
	AdditionalInfo string `gorm:"type:varchar(500)" json:"additionalInfo"`

	// Comment foreign key used for repleis
	//
	// required: false
	CommentId *uint `json:"commentId"`

	// Comment's replies
	Replies []Comment `json:"replies"`
}
