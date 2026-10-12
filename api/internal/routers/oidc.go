package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/oidc"

	"github.com/go-chi/chi/v5"
)

// BuildOidcRouter mounts the relying-party flow at /api/oidc.
//
// The static segments (link, exchange, connections) are declared as their own
// paths rather than under {name}, so chi's static-beats-param resolution stays
// unambiguous. The command validator additionally rejects those words as provider
// slugs, so no configured provider can shadow a route from the other direction.
//
// /link/{name}/launch is the one route that sits UNDER a param segment. It needs
// no reserved word: it is a segment deeper than /link/{name}, so the two cannot
// collide however a provider is named.
func BuildOidcRouter() *chi.Mux {
	router := chi.NewRouter()

	// Unauthenticated: this IS the sign-in path.
	router.Get("/{name}/login", oidc.Login)
	router.Get("/{name}/callback", oidc.Callback)
	router.Post("/exchange", handlers.OidcExchange)

	// Unauthenticated by design, and NOT a hole: the single-use launch handle in
	// the query is the authorization, minted seconds earlier by the authenticated
	// LinkStart below after checking app.account.update. It has to be reachable
	// without a bearer token precisely because the point is to get the EXTERNAL
	// BROWSER to start the session, so the binding cookie has somewhere to land.
	// Sits a segment deeper than /link/{name}, so there is no resolution overlap.
	router.Get("/link/{name}/launch", oidc.LinkLaunch)

	// Authenticated: connecting a provider to an account that already exists. The
	// session is what proves identity here, so nothing has to be inferred from a
	// claim -- which is why linkByUsername can safely default to off.
	router.Group(func(authenticated chi.Router) {
		authenticated.Use(middleware.UnifiedAuthMiddleware)

		authenticated.Get("/link/{name}", oidc.LinkStart)
		authenticated.Get("/connections", handlers.GetOidcConnections)
		authenticated.Delete("/connections/{name}", handlers.DeleteOidcConnection)
	})

	return router
}

// BuildOidcProviderRouter mounts administrator CRUD at /api/oidcProvider.
func BuildOidcProviderRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.UnifiedAuthMiddleware)

	router.Get("/{oidcProviderId}", handlers.GetOidcProviderById)
	router.Put("/{oidcProviderId}", handlers.UpdateOidcProvider)
	router.Delete("/{oidcProviderId}", handlers.DeleteOidcProvider)
	router.Post("/getPagedOidcProviders", handlers.GetPagedOidcProviders)
	router.Post("/", handlers.CreateOidcProvider)

	return router
}
