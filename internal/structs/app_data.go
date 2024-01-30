package structs

import "receipt-wrangler/api/internal/models"

type AppData struct {
	TokenPair
	Groups          []models.Group        `json:"groups"`
	UserPreferences models.UserPrefernces `json:"userPreferences"`
	Users           []UserView            `json:"users"`
	Claims          Claims                `json:"claims"`
}
