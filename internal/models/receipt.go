package models

type Receipt struct {
	BaseModel
	Name          string `gorm:"not null" json:"name"`
	ImgPath       string `json:"-"`
	PaidByUserID  uint
	PaidByUser    User `json:"-"`
	OwnedByUserID uint
	OwnedByUser   User `json:"-"`
}
