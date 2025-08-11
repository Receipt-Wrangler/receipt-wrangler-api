package models

type GroupReceiptSettings struct {
	BaseModel
	GroupId               uint `gorm:"not null;unique" json:"groupId"`
	HideImages            bool `gorm:"not null;default:false" json:"hideImages"`
	HideReceiptCategories bool `gorm:"not null;default:false" json:"hideReceiptCategories"`
	HideReceiptTags       bool `gorm:"not null;default:false" json:"hideReceiptTags"`
	HideItemCategories    bool `gorm:"not null;default:false" json:"hideItemCategories"`
	HideItemTags          bool `gorm:"not null;default:false" json:"hideItemTags"`
	HideShareCategories   bool `gorm:"not null;default:false" json:"hideShareCategories"`
	HideShareTags         bool `gorm:"not null;default:false" json:"hideShareTags"`
	HideComments          bool `gorm:"not null;default:false" json:"hideComments"`
}
