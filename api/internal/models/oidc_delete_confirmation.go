package models

import "time"

// OidcDeleteConfirmation is what an OIDC-only account presents to delete itself,
// in place of the password it does not have. The callback of a delete
// confirmation flow mints one after the account's own identity signed in at the
// provider again, and POST /user/deleteAccount spends it.
//
// It is its own table rather than a kind of OidcExchangeCode on purpose: an
// exchange code buys a session, and "delete my account" and "give me a session"
// must never share a redemption path. Hashed at rest, single-use and
// short-lived, like every other OIDC bearer secret. It is also useless on its
// own -- redeeming it needs a signed-in session of the same account.
type OidcDeleteConfirmation struct {
	BaseModel

	// CodeHash is sha256(code); the raw code only ever exists in the callback's
	// redirect and the client's delete request.
	CodeHash string `gorm:"not null;uniqueIndex;size:64" json:"-"`

	// UserId is the only account this code can delete.
	UserId uint `gorm:"not null;index" json:"-"`

	User *User `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE" json:"-"`

	Used      bool      `gorm:"not null;default:false;index" json:"-"`
	ExpiresAt time.Time `gorm:"index" json:"-"`
}
