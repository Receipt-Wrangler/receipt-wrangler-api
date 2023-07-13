package routers

import (
	"receipt-wrangler/api/internal/handlers"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildFeatureConfigRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	featureConfigRouter := chi.NewRouter()

	// swagger:route Get /featureConfig FeatureConfig featureConfig
	//
	// Get feature config
	//
	// This will get the server's feature config
	//
	//     Produces:
	//     - application/json
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
	featureConfigRouter.Get("/", handlers.GetFeatureConfig)

	return featureConfigRouter
}
