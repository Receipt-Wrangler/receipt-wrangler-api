package repositories

import (
	"receipt-wrangler/api/internal/models"
	"time"

	"gorm.io/gorm"
)

type OidcSessionRepository struct {
	BaseRepository
}

func NewOidcSessionRepository(tx *gorm.DB) OidcSessionRepository {
	repository := OidcSessionRepository{}
	repository.BaseRepository = BaseRepository{
		DB: GetDB(),
		TX: tx,
	}

	return repository
}

func (repository OidcSessionRepository) CreateAuthSession(session *models.OidcAuthSession) error {
	return repository.GetDB().Create(session).Error
}

// ConsumeAuthSession atomically claims an unused, unexpired session and returns
// it.
//
// The used and expiry checks live in the UPDATE's WHERE clause so check-and-set
// is a single statement: two concurrent callbacks bearing the same state can
// never both observe it as unused, and a replayed state affects zero rows. That
// matters more here than for an ordinary token, because the caller runs this
// BEFORE the token exchange -- so a replay cannot even reach the identity
// provider, let alone mint a second session.
//
// Reported as (session, claimed, error): claimed=false means the state was
// unknown, already used, or expired. Those are deliberately indistinguishable to
// the caller.
func (repository OidcSessionRepository) ConsumeAuthSession(stateHash string) (models.OidcAuthSession, bool, error) {
	db := repository.GetDB()

	result := db.Model(&models.OidcAuthSession{}).
		Where("state_hash = ? AND used = ? AND expires_at > ?", stateHash, false, time.Now()).
		Update("used", true)
	if result.Error != nil {
		return models.OidcAuthSession{}, false, result.Error
	}

	if result.RowsAffected != 1 {
		return models.OidcAuthSession{}, false, nil
	}

	var session models.OidcAuthSession
	err := db.Model(&models.OidcAuthSession{}).Where("state_hash = ?", stateHash).First(&session).Error
	if err != nil {
		return models.OidcAuthSession{}, false, err
	}

	return session, true, nil
}

func (repository OidcSessionRepository) CreateExchangeCode(code *models.OidcExchangeCode) error {
	return repository.GetDB().Create(code).Error
}

// GetExchangeCode loads a live exchange code WITHOUT consuming it, so the caller
// can verify the PKCE proof first.
//
// Load-then-verify-then-burn is deliberate, and copied from the OAuth token
// endpoint: burning the code before checking the proof would let anyone who
// intercepted the redirect destroy a valid code out from under the real client
// just by presenting a wrong verifier.
func (repository OidcSessionRepository) GetExchangeCode(codeHash string) (models.OidcExchangeCode, error) {
	db := repository.GetDB()
	var code models.OidcExchangeCode

	err := db.Model(&models.OidcExchangeCode{}).
		Where("code_hash = ? AND used = ? AND expires_at > ?", codeHash, false, time.Now()).
		First(&code).Error

	return code, err
}

// ConsumeExchangeCode atomically burns a code, mirroring ConsumeAuthSession.
func (repository OidcSessionRepository) ConsumeExchangeCode(codeHash string) (bool, error) {
	result := repository.GetDB().
		Model(&models.OidcExchangeCode{}).
		Where("code_hash = ? AND used = ? AND expires_at > ?", codeHash, false, time.Now()).
		Update("used", true)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

func (repository OidcSessionRepository) CreateLinkLaunch(launch *models.OidcLinkLaunch) error {
	return repository.GetDB().Create(launch).Error
}

// ConsumeLinkLaunch atomically burns a mobile link launch handle and returns the
// row it claimed, mirroring ConsumeAuthSession.
//
// Unlike the exchange code there is nothing to verify first, so this consumes
// outright rather than loading and burning separately: the handle IS the proof,
// and the browser presenting it gets exactly one attempt.
//
// Reported as (launch, claimed, error): claimed=false means the handle was
// unknown, already used, or expired, which are deliberately indistinguishable.
func (repository OidcSessionRepository) ConsumeLinkLaunch(launchHash string) (models.OidcLinkLaunch, bool, error) {
	db := repository.GetDB()

	result := db.Model(&models.OidcLinkLaunch{}).
		Where("launch_hash = ? AND used = ? AND expires_at > ?", launchHash, false, time.Now()).
		Update("used", true)
	if result.Error != nil {
		return models.OidcLinkLaunch{}, false, result.Error
	}

	if result.RowsAffected != 1 {
		return models.OidcLinkLaunch{}, false, nil
	}

	var launch models.OidcLinkLaunch
	err := db.Model(&models.OidcLinkLaunch{}).Where("launch_hash = ?", launchHash).First(&launch).Error
	if err != nil {
		return models.OidcLinkLaunch{}, false, err
	}

	return launch, true, nil
}

// DeleteExpiredOidcSessions is the cleanup cron's worker. All three tables are
// already guarded by expires_at inside the consume statements' WHERE clauses, so
// this is hygiene rather than a security control.
func (repository OidcSessionRepository) DeleteExpiredOidcSessions() error {
	db := repository.GetDB()
	now := time.Now()

	err := db.Where("expires_at < ? OR used = ?", now, true).Delete(&models.OidcAuthSession{}).Error
	if err != nil {
		return err
	}

	err = db.Where("expires_at < ? OR used = ?", now, true).Delete(&models.OidcExchangeCode{}).Error
	if err != nil {
		return err
	}

	return db.Where("expires_at < ? OR used = ?", now, true).Delete(&models.OidcLinkLaunch{}).Error
}
