package models

import "time"

// OidcExchangeCode is the mobile handoff: a short-lived, single-use code handed
// to the app over its private-use URL scheme, which the app then trades for real
// tokens at POST /api/oidc/exchange.
//
// Tokens are deliberately NOT put in the redirect URL — a private-use scheme is
// unverifiable on Android, so any installed app can register it and read the
// callback. The code is bound to the PKCE challenge the app generated before the
// flow started, so an app that intercepts the redirect still cannot redeem it:
// it never had the verifier.
type OidcExchangeCode struct {
	BaseModel

	// CodeHash is sha256(code); the raw code only ever exists in the redirect URL
	// and the app's exchange request.
	CodeHash string `gorm:"not null;uniqueIndex;size:64" json:"-"`

	UserId uint `gorm:"not null" json:"-"`

	// CodeChallenge is the app's S256 challenge, verified with
	// utils.VerifyPkceS256 against the verifier the app presents at exchange time.
	CodeChallenge string `gorm:"not null" json:"-"`

	Used      bool      `gorm:"not null;default:false;index" json:"-"`
	ExpiresAt time.Time `gorm:"index" json:"-"`
}
