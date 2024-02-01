package structs

import "receipt-wrangler/api/internal/models"

type AppData struct {
	Claims Claims         `json:"claims"`
	Groups []models.Group `json:"groups"`
	TokenPair
	UserPreferences models.UserPrefernces `json:"userPreferences"`
	Users           []UserView            `json:"users"`
	FeatureConfig   FeatureConfig         `json:"featureConfig"`
}
