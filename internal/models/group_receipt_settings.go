package models

type GroupReceiptSettings struct {
	BaseModel
	GroupId               uint `gorm:"not null;unique" json:"groupId"`
	HideImages            bool `json:"hideImages"`
	HideReceiptCategories bool `json:"hideReceiptCategories"`
	HideReceiptTags       bool `json:"hideReceiptTags"`
	HideShareCategories   bool `json:"hideShareCategories"`
	HideShareTags         bool `json:"hideShareTags"`
	HideComments          bool `json:"hideComments"`
}
