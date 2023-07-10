package models

// User in the system
//
// swagger:model
type User struct {
	BaseModel

	// Default avatar color
	//
	// required: false
	DefaultAvatarColor string `json:"defaultAvatarColor"`

	// Display name
	//
	// required: true
	DisplayName string `json:"displayName"`

	// Is dummy user
	//
	// required: true
	IsDummyUser bool `json:"isDummyUser"`

	// User's password
	//
	// required true
	Password string `gorm:"not null"`

	// User's username used to login
	//
	// required: true
	Username string `gorm:"not null; uniqueIndex"`

	// User's role in the system
	//
	// required: true
	UserRole UserRole `gorm:"default:'USER'" json:"userRole"`
}
