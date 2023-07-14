package models

// Category to relate receipts to
//
// swagger:model
type Category struct {
	BaseModel

	// Name of the category
	//
	// required: ture
	Name string `gorm:"not null; uniqueIndex" json:"name"`
}
