// Package oidc implements the OpenID Connect relying party: the side of the
// protocol that sends users to an external identity provider and verifies who
// comes back.
//
// It is the mirror image of internal/oauth, which is an OAuth 2.1 *authorization
// server* for the MCP integration. The two share no flow, only two primitives in
// internal/utils (PKCE verification and URL-safe random tokens).
//
// The API is a confidential client and owns the whole exchange:
//
//	discovery -> authorization code + PKCE S256 (state, nonce, browser binding)
//	-> code exchange with the client secret -> ID token verification
//	-> identity resolution -> Receipt Wrangler session
//
// No client -- desktop or mobile -- ever handles an identity provider's token or
// this deployment's client secret.
package oidc

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// discoveryTimeout bounds a discovery or JWKS fetch. The identity provider is an
// operator-supplied host, so a hung or hostile one must not be able to wedge a
// request thread indefinitely.
const discoveryTimeout = 10 * time.Second

type cacheEntry struct {
	provider  *oidc.Provider
	issuerUrl string
	updatedAt time.Time
}

var (
	providerCacheMutex sync.RWMutex
	providerCache      = map[uint]cacheEntry{}
)

// GetProvider returns a discovered *oidc.Provider for a configured row, caching
// it across requests.
//
// Caching is not an optimization here, it is the design. oidc.NewProvider makes a
// network round trip on EVERY call, and the *oidc.Provider is what owns the
// cached JWKS key set (it lazily builds one RemoteKeySet behind its own mutex and
// reuses it for every Verifier). Rebuilding per login would therefore refetch the
// signing keys on every login.
//
// A cache hit is valid only when both the issuer URL and the row's UpdatedAt
// match what was just read from the database. Since GORM bumps UpdatedAt on every
// write and the row is read fresh on each request, an administrator's edit
// invalidates the entry BY CONSTRUCTION -- including on other replicas, which
// never see the explicit InvalidateProvider call. That is why this needs no
// cross-process invalidation, and nobody should add any.
func GetProvider(providerRow models.OidcProvider) (*oidc.Provider, error) {
	providerCacheMutex.RLock()
	entry, ok := providerCache[providerRow.ID]
	providerCacheMutex.RUnlock()

	if ok && entry.issuerUrl == providerRow.IssuerUrl && entry.updatedAt.Equal(providerRow.UpdatedAt) {
		return entry.provider, nil
	}

	// Discovery runs OUTSIDE the write lock: it is a network call, and holding the
	// lock across it would serialize every login behind the slowest identity
	// provider. A cold-start race may discover twice; oidc.NewProvider is
	// idempotent, so last writer wins and both callers get a usable provider.
	//
	// The context is a fresh Background one with a timeout, NEVER the request's:
	// oidc.NewProvider stores the context's http.Client inside the Provider for its
	// later JWKS fetches, so a request-scoped context cancelled mid-discovery would
	// poison the cached entry for every subsequent request.
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	ctx = oidc.ClientContext(ctx, &http.Client{Timeout: discoveryTimeout})

	discovered, err := oidc.NewProvider(ctx, providerRow.IssuerUrl)
	if err != nil {
		return nil, err
	}

	providerCacheMutex.Lock()
	providerCache[providerRow.ID] = cacheEntry{
		provider:  discovered,
		issuerUrl: providerRow.IssuerUrl,
		updatedAt: providerRow.UpdatedAt,
	}
	providerCacheMutex.Unlock()

	return discovered, nil
}

// InvalidateProvider drops a cached entry. The UpdatedAt fingerprint already
// makes this unnecessary for correctness; it just frees the memory promptly on
// the process that handled the write.
func InvalidateProvider(id uint) {
	providerCacheMutex.Lock()
	delete(providerCache, id)
	providerCacheMutex.Unlock()
}

// ClearProviderCacheForTests empties the cache.
//
// The test database is truncated between cases and reuses row ids, so a cache
// keyed by id would serve one test's provider to another. Same hazard, and same
// remedy, as ClearRolePermissionCacheForTests.
func ClearProviderCacheForTests() {
	providerCacheMutex.Lock()
	providerCache = map[uint]cacheEntry{}
	providerCacheMutex.Unlock()
}

// buildOauthConfig assembles the oauth2 config for one provider. The client
// secret is decrypted here and never stored on a longer-lived structure.
func buildOauthConfig(
	providerRow models.OidcProvider,
	discovered *oidc.Provider,
	redirectUri string,
) (*oauth2.Config, error) {
	secret, err := repositories.NewOidcProviderRepository(nil).GetDecryptedClientSecret(providerRow)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     providerRow.ClientId,
		ClientSecret: secret,
		Endpoint:     discovered.Endpoint(),
		RedirectURL:  redirectUri,
		Scopes:       splitScope(providerRow.Scope),
	}, nil
}
