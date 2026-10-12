package structs

import "time"

// OidcProviderView is the read model for a configured provider.
//
// It deliberately omits the client secret. models.OidcProvider already carries
// `json:"-"` on that field, but this type is the belt to that suspenders: the
// model is trivially easy to marshal by accident (GetSystemSettings marshals its
// model directly), and a leaked client secret lets anyone impersonate this
// deployment to the identity provider.
type OidcProviderView struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	DisplayName       string `json:"displayName"`
	IssuerUrl         string `json:"issuerUrl"`
	ClientId          string `json:"clientId"`
	Scope             string `json:"scope"`
	AllowProvisioning bool   `json:"allowProvisioning"`
	LinkByUsername    bool   `json:"linkByUsername"`
	Enabled           bool   `json:"enabled"`
	// HasClientSecret lets the edit form say "a secret is configured" without ever
	// sending it, so leaving the field blank can safely mean "keep the stored one".
	HasClientSecret bool `json:"hasClientSecret"`
	// RedirectUri is computed from the server public URL, shown read-only so an
	// administrator can paste the exact string the identity provider must be
	// configured with. Providers match it exactly; a near miss fails at the IdP.
	RedirectUri string    `json:"redirectUri"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// OidcProviderSummary is the public, unauthenticated shape published on the
// feature config so a login screen can render its buttons. It carries only what
// is needed to name a provider and hit its login URL -- never the issuer, the
// client id, or anything else.
type OidcProviderSummary struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// OidcConnectionView is one row on the profile page's Connected Accounts list.
type OidcConnectionView struct {
	ProviderName        string `json:"providerName"`
	ProviderDisplayName string `json:"providerDisplayName"`
	PreferredUsername   string `json:"preferredUsername"`
	Email               string `json:"email"`
	// ProvisionedUser marks a connection that created the local account. Such an
	// account has no usable password, so the client hides Unlink when it is the
	// last one -- the server refuses it either way.
	ProvisionedUser bool       `json:"provisionedUser"`
	LinkedAt        time.Time  `json:"linkedAt"`
	LastLoginAt     *time.Time `json:"lastLoginAt"`
}

// OidcLinkStartView is the response to a MOBILE "connect account" start.
//
// The desktop gets a 302 straight to the identity provider, but the mobile app
// authenticates with a bearer token and the external user agent RFC 8252 demands
// cannot carry one. So the app starts the flow as an ordinary authenticated API
// call and opens the URL this carries itself.
//
// LaunchUrl points back at THIS server, not at the identity provider, and that
// indirection is the security property. It used to carry the authorization URL
// directly, justified as "possessing it grants nothing the bearer token did not
// already" -- which was wrong. That URL carries the state, identity providers
// leak it by Referer from their own login pages, and the callback it leads to had
// no way to tell the app's browser from anyone else's. Whoever held it could
// complete the flow as themselves and have their identity grafted onto this
// caller's account.
//
// Now the app opens a launch URL here instead. The external browser follows it,
// so the SERVER gets to mint the session and set a binding cookie on the agent
// that will return with the callback -- the same defense every other leg has.
type OidcLinkStartView struct {
	LaunchUrl string `json:"launchUrl"`
}

// OidcFlowError reports a failure from a flow start that answers with JSON
// rather than a redirect. It carries the same small fixed vocabulary of codes
// the redirect form puts in its query string, so a client maps both with one
// table -- and, as everywhere else in this package, never an identity
// provider's own error text.
type OidcFlowError struct {
	ErrorCode string `json:"errorCode"`
}
