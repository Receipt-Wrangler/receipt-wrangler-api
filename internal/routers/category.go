package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildCategoryRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	categoryRouter := chi.NewRouter()

	categoryRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route GET /category/ Category category
	//
	// Get all categories
	//
	// This will return all categories in the system
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
	categoryRouter.Get("/", handlers.GetAllCategories)

	return categoryRouter
}
