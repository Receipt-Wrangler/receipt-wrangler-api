package models

type Notification struct {
	BaseModel
	Type NotificationType `gorm:"not null; uniqueIndex" json:"type"`
	// Bodies will contain parsable things for the frontend ${userId:1} for example. This way we can get the data on the client to resolve what that is
	// ${userId:1} added a receipt in ${groupId:1}
	Body   string `gorm:"not null"; json:"body"`
	UserId uint   `json:"paidByUserId"`
	User   User   `json:"-"`
}
