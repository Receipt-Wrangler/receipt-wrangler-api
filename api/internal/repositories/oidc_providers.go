package repositories

import (
	"errors"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"strings"

	"gorm.io/gorm"
)

type OidcProviderRepository struct {
	BaseRepository
}

func NewOidcProviderRepository(tx *gorm.DB) OidcProviderRepository {
	repository := OidcProviderRepository{}
	repository.BaseRepository = BaseRepository{
		DB: GetDB(),
		TX: tx,
	}

	return repository
}

func (repository OidcProviderRepository) GetPagedOidcProviders(
	pagedRequestCommand commands.PagedRequestCommand,
) ([]models.OidcProvider, int64, error) {
	db := repository.GetDB()
	var providers []models.OidcProvider

	err := repository.validateOrderBy(pagedRequestCommand.OrderBy)
	if err != nil {
		return providers, 0, err
	}

	query := repository.Sort(db, pagedRequestCommand.OrderBy, pagedRequestCommand.SortDirection)
	query = query.Scopes(repository.Paginate(pagedRequestCommand.Page, pagedRequestCommand.PageSize))

	err = query.Model(&models.OidcProvider{}).Find(&providers).Error
	if err != nil {
		return nil, 0, err
	}

	var count int64
	err = db.Model(&models.OidcProvider{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	return providers, count, nil
}

func (repository OidcProviderRepository) GetOidcProviderById(id uint) (models.OidcProvider, error) {
	db := repository.GetDB()
	var provider models.OidcProvider

	err := db.Model(&models.OidcProvider{}).Where("id = ?", id).First(&provider).Error

	return provider, err
}

// GetEnabledOidcProviderByName resolves the provider a request's URL slug names.
// Disabled providers are deliberately indistinguishable from missing ones, so
// turning a provider off closes every one of its routes at once.
func (repository OidcProviderRepository) GetEnabledOidcProviderByName(name string) (models.OidcProvider, error) {
	db := repository.GetDB()
	var provider models.OidcProvider

	err := db.Model(&models.OidcProvider{}).
		Where("name = ? AND enabled = ?", strings.ToLower(strings.TrimSpace(name)), true).
		First(&provider).Error

	return provider, err
}

// GetAllEnabledOidcProviders backs the public feature config's provider list and
// the profile page's "not yet connected" list.
func (repository OidcProviderRepository) GetAllEnabledOidcProviders() ([]models.OidcProvider, error) {
	db := repository.GetDB()
	providers := make([]models.OidcProvider, 0)

	err := db.Model(&models.OidcProvider{}).
		Where("enabled = ?", true).
		Order("display_name asc").
		Find(&providers).Error

	return providers, err
}

func (repository OidcProviderRepository) GetOidcProviderByName(name string) (models.OidcProvider, error) {
	db := repository.GetDB()
	var provider models.OidcProvider

	err := db.Model(&models.OidcProvider{}).
		Where("name = ?", strings.ToLower(strings.TrimSpace(name))).
		First(&provider).Error

	return provider, err
}

func (repository OidcProviderRepository) CreateOidcProvider(
	command commands.UpsertOidcProviderCommand,
	createdBy *uint,
) (models.OidcProvider, error) {
	db := repository.GetDB()

	secret := ""
	if command.ClientSecret != nil {
		secret = *command.ClientSecret
	}

	encryptedSecret, err := encryptOidcClientSecret(secret)
	if err != nil {
		return models.OidcProvider{}, err
	}

	provider := models.OidcProvider{
		Name:              command.Name,
		DisplayName:       command.DisplayName,
		IssuerUrl:         command.IssuerUrl,
		ClientId:          command.ClientId,
		ClientSecret:      encryptedSecret,
		Scope:             command.Scope,
		AllowProvisioning: command.AllowProvisioning,
		LinkByUsername:    command.LinkByUsername,
		Enabled:           command.Enabled,
	}
	provider.CreatedBy = createdBy

	err = db.Create(&provider).Error
	if err != nil {
		return models.OidcProvider{}, err
	}

	return provider, nil
}

// UpdateOidcProvider writes the provider with the MAP form of Updates.
//
// GORM's struct form skips zero values, which would silently ignore turning
// allow_provisioning, link_by_username or enabled OFF -- and each of those is a
// security toggle, so an admin revoking one would believe they had. Same reason
// UpdateCustomField and UpdateAppRole use the map form.
//
// client_secret is only in the map when the request actually supplied one, so an
// update that leaves the field blank keeps the stored ciphertext byte for byte.
func (repository OidcProviderRepository) UpdateOidcProvider(
	id uint,
	command commands.UpsertOidcProviderCommand,
) (models.OidcProvider, error) {
	db := repository.GetDB()

	updates := map[string]interface{}{
		"display_name":       command.DisplayName,
		"issuer_url":         command.IssuerUrl,
		"client_id":          command.ClientId,
		"scope":              command.Scope,
		"allow_provisioning": command.AllowProvisioning,
		"link_by_username":   command.LinkByUsername,
		"enabled":            command.Enabled,
	}

	if command.ClientSecret != nil {
		encryptedSecret, err := encryptOidcClientSecret(*command.ClientSecret)
		if err != nil {
			return models.OidcProvider{}, err
		}

		updates["client_secret"] = encryptedSecret
	}

	err := db.Model(&models.OidcProvider{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return models.OidcProvider{}, err
	}

	return repository.GetOidcProviderById(id)
}

// DeleteOidcProvider removes the provider and everything hanging off it in one
// transaction. The cascades are explicit rather than FK-driven, matching every
// other cascade in this codebase: SQLite does not enforce foreign keys unless
// the pragma is set per connection, so an orphaned identity row would otherwise
// survive and could be re-adopted if the slug were later reused.
func (repository OidcProviderRepository) DeleteOidcProvider(id uint) error {
	db := repository.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("oidc_provider_id = ?", id).Delete(&models.OidcIdentity{}).Error
		if err != nil {
			return err
		}

		err = tx.Where("oidc_provider_id = ?", id).Delete(&models.OidcAuthSession{}).Error
		if err != nil {
			return err
		}

		return tx.Where("id = ?", id).Delete(&models.OidcProvider{}).Error
	})
}

// GetDecryptedClientSecret is the only place the client secret is decrypted.
func (repository OidcProviderRepository) GetDecryptedClientSecret(provider models.OidcProvider) (string, error) {
	if len(provider.ClientSecret) == 0 {
		return "", errors.New("oidc provider has no client secret")
	}

	return utils.DecryptB64EncodedData(env.GetEncryptionKey(), provider.ClientSecret)
}

func encryptOidcClientSecret(secret string) (string, error) {
	return utils.EncryptAndEncodeToBase64(env.GetEncryptionKey(), secret)
}

func (repository OidcProviderRepository) validateOrderBy(orderBy string) error {
	switch orderBy {
	case "name", "display_name", "issuer_url", "enabled", "created_at":
		return nil
	}

	return errors.New("invalid orderBy")
}
