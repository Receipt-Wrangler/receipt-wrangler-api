package routers

import (
	"receipt-wrangler/api/internal/handlers"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildFeatureConfigRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	featureConfigRouter := chi.NewRouter()
	featureConfigRouter.Get("/", handlers.GetFeatureConfig)

	return featureConfigRouter
}
