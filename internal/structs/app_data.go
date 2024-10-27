package structs

import "receipt-wrangler/api/internal/models"

type AppData struct {
	TokenPair
	About                        About                         `json:"about"`
	Claims                       Claims                        `json:"claims"`
	Groups                       []models.Group                `json:"groups"`
	CurrencyDisplay              string                        `json:"currencyDisplay"`
	CurrencyThousandthsSeparator models.CurrencySeparator      `json:"currencyThousandthsSeparator"`
	CurrencyDecimalSeparator     models.CurrencySeparator      `json:"currencyDecimalSeparator"`
	CurrencySymbolPosition       models.CurrencySymbolPosition `json:"currencySymbolPosition"`
	CurrencyHideDecimalPlaces    bool                          `json:"currencyHideDecimalPlaces"`
	Categories                   []models.Category             `json:"categories"`
	Tags                         []models.Tag                  `json:"tags"`
	UserPreferences              models.UserPrefernces         `json:"userPreferences"`
	Users                        []UserView                    `json:"users"`
	FeatureConfig                FeatureConfig                 `json:"featureConfig"`
	Icons                        []Icon                        `json:"icons"`
}
