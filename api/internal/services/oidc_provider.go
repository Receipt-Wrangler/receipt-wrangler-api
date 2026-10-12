package services

import (
	"errors"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"

	"gorm.io/gorm"
)

var (
	// ErrOidcProviderNameTaken means another provider already owns the slug.
	ErrOidcProviderNameTaken = errors.New("an OIDC provider with this name already exists")

	// ErrServerPublicUrlRequired means the deployment has no public origin
	// configured, so there is no redirect URI to register with the identity
	// provider. Enabling a provider without it would silently produce a localhost
	// redirect that fails at the IdP.
	ErrServerPublicUrlRequired = errors.New("a server public URL must be configured in system settings before enabling an OIDC provider")
)

type OidcProviderService struct {
	BaseService
}

func NewOidcProviderService(tx *gorm.DB) OidcProviderService {
	service := OidcProviderService{}
	service.BaseService = BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}

	return service
}

func (service OidcProviderService) CreateOidcProvider(
	command commands.UpsertOidcProviderCommand,
	createdBy *uint,
) (structs.OidcProviderView, error) {
	repository := repositories.NewOidcProviderRepository(service.TX)

	err := service.validateCrossSettingRules(command)
	if err != nil {
		return structs.OidcProviderView{}, err
	}

	_, err = repository.GetOidcProviderByName(command.Name)
	if err == nil {
		return structs.OidcProviderView{}, ErrOidcProviderNameTaken
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return structs.OidcProviderView{}, err
	}

	provider, err := repository.CreateOidcProvider(command, createdBy)
	if err != nil {
		return structs.OidcProviderView{}, err
	}

	return BuildOidcProviderView(provider), nil
}

func (service OidcProviderService) UpdateOidcProvider(
	id uint,
	command commands.UpsertOidcProviderCommand,
) (structs.OidcProviderView, error) {
	repository := repositories.NewOidcProviderRepository(service.TX)

	err := service.validateCrossSettingRules(command)
	if err != nil {
		return structs.OidcProviderView{}, err
	}

	provider, err := repository.UpdateOidcProvider(id, command)
	if err != nil {
		return structs.OidcProviderView{}, err
	}

	return BuildOidcProviderView(provider), nil
}

// validateCrossSettingRules holds the checks a pure command validator cannot
// make because they read other settings.
func (service OidcProviderService) validateCrossSettingRules(command commands.UpsertOidcProviderCommand) error {
	if command.Enabled && !IsServerPublicUrlConfigured() {
		return ErrServerPublicUrlRequired
	}

	return nil
}

// BuildOidcProviderView maps a stored provider to its read model, computing the
// redirect URI from the live server public URL.
func BuildOidcProviderView(provider models.OidcProvider) structs.OidcProviderView {
	return structs.OidcProviderView{
		ID:                provider.ID,
		Name:              provider.Name,
		DisplayName:       provider.DisplayName,
		IssuerUrl:         provider.IssuerUrl,
		ClientId:          provider.ClientId,
		Scope:             provider.Scope,
		AllowProvisioning: provider.AllowProvisioning,
		LinkByUsername:    provider.LinkByUsername,
		Enabled:           provider.Enabled,
		HasClientSecret:   len(provider.ClientSecret) > 0,
		RedirectUri:       BuildOidcRedirectUri(provider.Name),
		CreatedAt:         provider.CreatedAt,
		UpdatedAt:         provider.UpdatedAt,
	}
}

// BuildOidcProviderViews maps a slice, always returning a non-nil result so it
// serializes as [] rather than null.
func BuildOidcProviderViews(providers []models.OidcProvider) []structs.OidcProviderView {
	views := make([]structs.OidcProviderView, 0, len(providers))

	for _, provider := range providers {
		views = append(views, BuildOidcProviderView(provider))
	}

	return views
}

// BuildOidcConnectionViews maps a user's identities for the profile page.
func BuildOidcConnectionViews(identities []models.OidcIdentity) []structs.OidcConnectionView {
	views := make([]structs.OidcConnectionView, 0, len(identities))

	for _, identity := range identities {
		view := structs.OidcConnectionView{
			PreferredUsername: identity.PreferredUsername,
			Email:             identity.Email,
			ProvisionedUser:   identity.ProvisionedUser,
			LinkedAt:          identity.CreatedAt,
			LastLoginAt:       identity.LastLoginAt,
		}

		if identity.OidcProvider != nil {
			view.ProviderName = identity.OidcProvider.Name
			view.ProviderDisplayName = identity.OidcProvider.DisplayName
		}

		views = append(views, view)
	}

	return views
}
