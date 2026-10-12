package repositories

import (
	"receipt-wrangler/api/internal/models"
	"time"

	"gorm.io/gorm"
)

type OidcIdentityRepository struct {
	BaseRepository
}

func NewOidcIdentityRepository(tx *gorm.DB) OidcIdentityRepository {
	repository := OidcIdentityRepository{}
	repository.BaseRepository = BaseRepository{
		DB: GetDB(),
		TX: tx,
	}

	return repository
}

// GetIdentityBySubject is the identity anchor lookup. Every login after the very
// first resolves through exactly this query and consults no other claim, which is
// what makes a username change (or an IdP recycling a released name) unable to
// re-point an existing account.
func (repository OidcIdentityRepository) GetIdentityBySubject(providerId uint, subject string) (models.OidcIdentity, error) {
	db := repository.GetDB()
	var identity models.OidcIdentity

	err := db.Model(&models.OidcIdentity{}).
		Where("oidc_provider_id = ? AND subject = ?", providerId, subject).
		First(&identity).Error

	return identity, err
}

func (repository OidcIdentityRepository) GetIdentityForUser(providerId uint, userId uint) (models.OidcIdentity, error) {
	db := repository.GetDB()
	var identity models.OidcIdentity

	err := db.Model(&models.OidcIdentity{}).
		Where("oidc_provider_id = ? AND user_id = ?", providerId, userId).
		First(&identity).Error

	return identity, err
}

// GetIdentitiesForUser returns every provider the user has connected. The slice
// is always non-nil so it serializes as [] rather than null.
func (repository OidcIdentityRepository) GetIdentitiesForUser(userId uint) ([]models.OidcIdentity, error) {
	db := repository.GetDB()
	identities := make([]models.OidcIdentity, 0)

	err := db.Model(&models.OidcIdentity{}).
		Preload("OidcProvider").
		Where("user_id = ?", userId).
		Order("id asc").
		Find(&identities).Error

	return identities, err
}

func (repository OidcIdentityRepository) CountIdentitiesForUser(userId uint) (int64, error) {
	db := repository.GetDB()
	var count int64

	err := db.Model(&models.OidcIdentity{}).Where("user_id = ?", userId).Count(&count).Error

	return count, err
}

// CountProvisionedIdentitiesForUser counts the account's identities that are
// flagged as belonging to a passwordless account.
//
// The flag is held on EVERY identity of such an account, not just the one that
// created it -- see the invariant on models.OidcIdentity.ProvisionedUser -- so
// any non-zero count means the account still has no password of its own.
func (repository OidcIdentityRepository) CountProvisionedIdentitiesForUser(userId uint) (int64, error) {
	db := repository.GetDB()
	var count int64

	err := db.Model(&models.OidcIdentity{}).
		Where("user_id = ? AND provisioned_user = ?", userId, true).
		Count(&count).Error

	return count, err
}

// SetProvisionedFlagForUser re-asserts the passwordless fact across every
// identity the account holds, restoring the invariant on rows that predate it.
func (repository OidcIdentityRepository) SetProvisionedFlagForUser(userId uint) error {
	return repository.GetDB().
		Model(&models.OidcIdentity{}).
		Where("user_id = ?", userId).
		Update("provisioned_user", true).Error
}

// ClearProvisionedFlagForUser records that the account now has a real password,
// across every identity it holds. Called when a password is actually set, which
// is what makes the lockout guard's advice -- "ask an administrator to set one"
// -- something that works rather than a dead end.
func (repository OidcIdentityRepository) ClearProvisionedFlagForUser(userId uint) error {
	return repository.GetDB().
		Model(&models.OidcIdentity{}).
		Where("user_id = ?", userId).
		Update("provisioned_user", false).Error
}

func (repository OidcIdentityRepository) CreateIdentity(identity *models.OidcIdentity) error {
	return repository.GetDB().Create(identity).Error
}

// TouchIdentity refreshes the display-only claims and the last-login stamp. These
// values are never used to resolve an identity -- only to render the Connected
// Accounts row -- so a change here can never affect who a login lands on.
func (repository OidcIdentityRepository) TouchIdentity(id uint, preferredUsername string, email string) error {
	now := time.Now()

	return repository.GetDB().
		Model(&models.OidcIdentity{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"preferred_username": preferredUsername,
			"email":              email,
			"last_login_at":      &now,
		}).Error
}

func (repository OidcIdentityRepository) DeleteIdentity(id uint, userId uint) error {
	// Scoped by user id as well as row id so a caller can only ever unlink their
	// own identity, even if the id came from somewhere it should not have.
	return repository.GetDB().
		Where("id = ? AND user_id = ?", id, userId).
		Delete(&models.OidcIdentity{}).Error
}
