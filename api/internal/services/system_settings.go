package services

import (
	"net/url"
	"strings"

	"gorm.io/gorm"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

// loginQrDeepLinkBase is the App Link / Universal Link entry point that the
// login QR encodes. It is a fixed, project-owned domain (the same one baked
// into the published mobile app and its hosted assetlinks.json /
// apple-app-site-association files) — it is NOT the self-hoster's own domain.
// Keep this in sync with the mobile App Link config and the hosted files.
const loginQrDeepLinkBase = "https://receiptwrangler.io/app/setup"

// BuildLoginQrUrl composes the deep link the desktop login page renders as a
// QR. The server URL rides in the fragment (#url=...) so it never reaches
// receiptwrangler.io's server logs in the app-not-installed web fallback.
func BuildLoginQrUrl(serverUrl string) string {
	return loginQrDeepLinkBase + "#url=" + url.QueryEscape(serverUrl)
}

type SystemSettingsService struct {
	BaseService
}

func NewSystemSettingsService(tx *gorm.DB) SystemSettingsService {
	service := SystemSettingsService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service SystemSettingsService) GetFeatureConfig() (structs.FeatureConfig, error) {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(service.TX)
	featureConfig := structs.FeatureConfig{
		// Initialized before anything can fail, and only ever overwritten on
		// success, so this can never serialize as null. See the field's doc comment.
		OidcProviders: []structs.OidcProviderSummary{},
	}

	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return structs.FeatureConfig{}, err
	}

	aiPoweredReceipts := systemSettings.ReceiptProcessingSettingsId != nil

	featureConfig.EnableLocalSignUp = systemSettings.EnableLocalSignUp
	featureConfig.AiPoweredReceipts = aiPoweredReceipts

	mobileServerUrl := strings.TrimSpace(systemSettings.MobileServerUrl)
	if systemSettings.ShowLoginQr && len(mobileServerUrl) > 0 {
		featureConfig.LoginQrUrl = BuildLoginQrUrl(mobileServerUrl)
	}

	providers, err := repositories.NewOidcProviderRepository(service.TX).GetAllEnabledOidcProviders()
	if err != nil {
		return structs.FeatureConfig{}, err
	}

	for _, provider := range providers {
		featureConfig.OidcProviders = append(featureConfig.OidcProviders, structs.OidcProviderSummary{
			Name:        provider.Name,
			DisplayName: provider.DisplayName,
		})
	}

	return featureConfig, nil
}
