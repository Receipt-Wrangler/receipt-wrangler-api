package models

import "time"

// OidcLinkLaunch is what makes the mobile "connect account" leg bindable.
//
// Every other OIDC start is a browser navigation, so the server can set a
// binding cookie on it and refuse a callback that does not carry it back. A
// mobile link start is not: the app calls it as an ordinary bearer-authenticated
// API request, because an external user agent cannot carry a bearer token. There
// is no browser in that exchange to hand a cookie to.
//
// So the mobile link start creates no auth session at all. It creates one of
// these instead and returns a launch URL, which the app opens in the external
// browser. That request DOES reach the browser, so it can mint the session,
// set the binding cookie and redirect on to the identity provider.
//
// The handle is the browser's authorization to start a link for this user, so it
// is hashed at rest, single-use and short-lived, exactly like the state and the
// exchange code. It never travels to the identity provider: a 302 does not make
// the redirecting URL the next request's Referer, which matters because leaking
// the authorization URL by Referer is the very attack this hop defeats.
type OidcLinkLaunch struct {
	BaseModel

	// LaunchHash is sha256(handle). The raw handle exists only in the authenticated
	// response body and in the app's memory.
	LaunchHash string `gorm:"not null;uniqueIndex;size:64" json:"-"`

	// UserId is the account the resulting link will attach to. It comes from the
	// authenticated caller of the link start, never from the launch request, which
	// is why that request can safely be unauthenticated.
	UserId uint `gorm:"not null" json:"-"`

	OidcProviderId uint `gorm:"not null;index" json:"-"`

	OidcProvider *OidcProvider `gorm:"foreignKey:OidcProviderId;constraint:OnDelete:CASCADE" json:"-"`
	User         *User         `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE" json:"-"`

	Used      bool      `gorm:"not null;default:false;index" json:"-"`
	ExpiresAt time.Time `gorm:"index" json:"-"`
}
