package models

// OidcProvider is an administrator-configured OpenID Connect identity provider
// that users can sign in with. The Go API acts as the OIDC **relying party**
// (a confidential client): discovery, PKCE, state, nonce, the code exchange and
// ID-token verification all happen server-side, so no client ever handles an
// IdP token or this client secret.
//
// Not to be confused with internal/oauth, which is the mirror image — an OAuth
// 2.1 *authorization server* for the MCP integration.
type OidcProvider struct {
	BaseModel

	// Name is the URL slug for this provider (/api/oidc/{name}/login). It is
	// immutable after creation because it is baked into the redirect URI already
	// registered at the IdP — a rename would silently break every future login.
	// A small set of route-shadowing slugs is reserved; see
	// commands.ReservedOidcProviderNames.
	Name string `gorm:"not null;uniqueIndex;size:64" json:"name"`

	// DisplayName is rendered to end users as "Log in with {{ displayName }}".
	DisplayName string `gorm:"not null" json:"displayName"`

	// IssuerUrl is the OIDC discovery base, e.g. https://accounts.google.com.
	// go-oidc validates that the discovery document's own "issuer" matches it.
	IssuerUrl string `gorm:"not null" json:"issuerUrl"`

	ClientId string `gorm:"not null" json:"clientId"`

	// ClientSecret is stored AES-256-GCM encrypted (utils.EncryptAndEncodeToBase64
	// with config.GetEncryptionKey), matching SystemEmail.Password and
	// ReceiptProcessingSettings.Key. `json:"-"` keeps it off the wire even if this
	// model is ever marshalled directly; structs.OidcProviderView is the read model
	// that callers actually receive.
	ClientSecret string `gorm:"not null" json:"-"`

	// Scope is the space-separated OIDC scope string. It must contain "openid";
	// without it the provider returns an OAuth response with no ID token and the
	// callback hard-fails.
	Scope string `gorm:"not null;default:'openid profile email'" json:"scope"`

	// AllowProvisioning creates a local account for an IdP identity we have never
	// seen. When false, an unknown identity is refused rather than provisioned.
	AllowProvisioning bool `gorm:"not null;default:false" json:"allowProvisioning"`

	// LinkByUsername lets a FIRST login (no link row yet) attach to an existing
	// local account whose username equals the IdP's preferred_username claim.
	//
	// Defaults to false, and deliberately so: OIDC does not guarantee
	// preferred_username is stable or unique, and several public IdPs (Twitch
	// notably) let users change their username and recycle released names — so an
	// always-on match is an account-takeover path against a local "admin". The safe
	// way to attach an existing account is the authenticated link flow, which
	// proves identity with a session instead of guessing from a claim.
	LinkByUsername bool `gorm:"not null;default:false" json:"linkByUsername"`

	// Enabled hides the provider's login button without discarding its
	// configuration. A disabled provider 404s on every OIDC route.
	//
	// Deliberately NO `default:true` tag: GORM skips zero-value fields on Create
	// when the column has a default, so a provider created with enabled=false
	// would silently come back enabled. Both create and update always write this
	// column explicitly, and defaulting an omitted value to false is the safe
	// direction anyway -- a provider nobody explicitly enabled should not be live.
	Enabled bool `gorm:"not null" json:"enabled"`
}
