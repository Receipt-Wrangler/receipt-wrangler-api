package models

import "time"

// OidcIdentity links a local User to one identity at one OidcProvider.
//
// This is the identity anchor for the whole feature. The link is keyed on the
// ID token's `sub` claim, which is the ONLY claim OIDC guarantees is both stable
// and unique within an issuer and never reassigned. Every login after the first
// is a lookup on (OidcProviderId, Subject) and consults no other claim — so a
// user renaming themselves at the IdP, or an IdP recycling a released username,
// can never re-point an existing link.
type OidcIdentity struct {
	BaseModel

	// Subject is the ID token's `sub`. Capped at 255 so the composite unique
	// index below stays inside MySQL's 3072-byte index limit; that is also why
	// PreferredUsername is deliberately NOT part of any index.
	Subject        string `gorm:"not null;size:255;uniqueIndex:idx_oidc_identity_provider_subject" json:"subject"`
	OidcProviderId uint   `gorm:"not null;uniqueIndex:idx_oidc_identity_provider_subject;uniqueIndex:idx_oidc_identity_provider_user" json:"oidcProviderId"`

	// UserId, together with the provider, is unique: one local user holds at most
	// one identity per provider, so the profile page can never show duplicates and
	// an unlink/relink cycle cannot leave a stale row that still logs in.
	UserId uint `gorm:"not null;index;uniqueIndex:idx_oidc_identity_provider_user" json:"userId"`

	OidcProvider *OidcProvider `gorm:"foreignKey:OidcProviderId;constraint:OnDelete:CASCADE" json:"-"`
	User         *User         `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE" json:"-"`

	// PreferredUsername and Email are the last-seen values of those claims, kept
	// for display on the Connected Accounts row only. They are refreshed on every
	// login and are NEVER used to resolve an identity once the link exists.
	PreferredUsername string `json:"preferredUsername"`
	Email             string `json:"email"`

	// ProvisionedUser records that THE ACCOUNT has no password of its own: it was
	// created by an OIDC provisioning login, which only ever gave it a random,
	// discarded password. Unlinking its last identity would strand it with no way
	// back in — see the unlink lockout guard in the OIDC service.
	//
	// INVARIANT: it is held identically on EVERY identity of such an account, and
	// on none of an account that has a real password. Three places maintain it —
	// provisionUser sets it, resolveLink inherits it when attaching a second
	// provider, and ClearProvisionedFlagForUser clears it across all of them when a
	// password is actually set.
	//
	// The name is narrower than the meaning, and deliberately unchanged: it is on
	// the wire as `provisionedUser`, required and non-nullable in the generated Dart
	// client, so renaming it would break already-released mobile builds for a
	// cosmetic gain.
	//
	// It reads like it belongs on User, and it originally meant the narrower "this
	// identity created the account", which is why it lives here. Holding it per
	// identity is what made the first version of the lockout guard wrong: the fact
	// died with the row that carried it. The invariant above is what repairs that
	// without widening the User model.
	ProvisionedUser bool `gorm:"not null;default:false" json:"provisionedUser"`

	LastLoginAt *time.Time `json:"lastLoginAt"`
}
