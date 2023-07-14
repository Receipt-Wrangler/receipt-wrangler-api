package models

// Notification
//
// swagger:model
type Notification struct {
	BaseModel

	// Notification type
	//
	// required: true
	Type NotificationType `gorm:"not null" json:"type"`
	// ${userId:1} added a receipt in ${groupId:1} // ${id:value.key:type}

	// Title
	//
	// required: true
	Title string `gorm:"not null;" json:"title"`

	// Notification body
	//
	// requried: true
	Body string `gorm:"not null;" json:"body"`

	// User foreign key
	//
	// required: true
	UserId uint `json:"userId"`

	User User `json:"-"`
}
