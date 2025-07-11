package models

type GroupReceiptSettings struct {
	BaseModel
	GroupId               uint `gorm:"not null;unique" json:"groupId"`
	HideImages            bool `gorm:"not null" json:"hideImages"`
	HideReceiptCategories bool `gorm:"not null" json:"hideReceiptCategories"`
	HideReceiptTags       bool `gorm:"not null" json:"hideReceiptTags"`
	HideShareCategories   bool `gorm:"not null" json:"hideShareCategories"`
	HideShareTags         bool `gorm:"not null" json:"hideShareTags"`
	HideComments          bool `gorm:"not null" json:"hideComments"`
}
