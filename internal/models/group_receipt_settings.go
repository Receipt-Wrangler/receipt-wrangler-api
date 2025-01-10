package models

type GroupReceiptSettings struct {
	BaseModel
	GroupId               uint `gorm:"not null;unique" json:"groupId"`
	HideImages            bool `gorm:"not null" json:"hideImages"`
	HideReceiptCategories bool `gorm:"not null" json:"hideReceiptCategories"`
	HideReceiptTags       bool `gorm:"not null" json:"hideReceiptTags"`
	HideItemCategories    bool `gorm:"not null" json:"hideItemCategories"`
	HideItemTags          bool `gorm:"not null" json:"hideItemTags"`
	HideComments          bool `gorm:"not null" json:"hideComments"`
}
