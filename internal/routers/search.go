package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildSearchRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	searchRouter := chi.NewRouter()

	searchRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route GET /search Search search
	//
	// Receipt Search
	//
	// This will search for receipts based on a search term
	//
	//     Consumes:
	//     - text/plain
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//       + name: searchTerm
	//         in: query
	//         description: search term
	//         required: true
	//         type: string
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	searchRouter.Get("/", handlers.Search)

	return searchRouter
}
