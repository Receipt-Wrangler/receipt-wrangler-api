package structs

import "receipt-wrangler/api/internal/models"

type AppData struct {
	Groups          []models.Group        `json:"groups"`
	Jwt             string                `json:"jwt"`
	RefreshToken    string                `json:"refreshToken"`
	UserPreferences models.UserPrefernces `json:"userPreferences"`
	Users           []UserView            `json:"users"`
	Claims          Claims                `json:"claims"`
}
