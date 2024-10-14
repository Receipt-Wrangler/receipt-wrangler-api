package structs

import "receipt-wrangler/api/internal/models"

type AppData struct {
	TokenPair
	Claims                       Claims                `json:"claims"`
	Groups                       []models.Group        `json:"groups"`
	CurrencyDisplay              string                `json:"currencyDisplay"`
	CurrencyThousandthsSeparator string                `json:"currencyThousandthsSeparator"`
	CurrencyDecimalSeparator     string                `json:"currencyDecimalSeparator"`
	CurrencySymbolPosition       string                `json:"currencySymbolPosition"`
	Categories                   []models.Category     `json:"categories"`
	Tags                         []models.Tag          `json:"tags"`
	UserPreferences              models.UserPrefernces `json:"userPreferences"`
	Users                        []UserView            `json:"users"`
	FeatureConfig                FeatureConfig         `json:"featureConfig"`
	Icons                        []Icon                `json:"icons"`
}
