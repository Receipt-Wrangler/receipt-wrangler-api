package models

// Tag to relate receipts to
//
// swagger:model
type Tag struct {
	BaseModel

	// Tag name
	//
	// required: true
	Name string `gorm:"not null; uniqueIndex" json:"name"`
}
