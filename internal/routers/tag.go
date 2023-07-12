package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildTagRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	tagRouter := chi.NewRouter()

	tagRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route GET /tag/ Tag tag
	//
	// Get all tags
	//
	// This will return all tags in the system
	//
	//
	//
	//     Produces:
	//     - application/json
	//
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
	tagRouter.Get("/", handlers.GetAllTags)

	return tagRouter
}
