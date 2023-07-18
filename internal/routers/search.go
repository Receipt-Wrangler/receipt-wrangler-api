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
	searchRouter.Get("/", handlers.Search)

	return searchRouter
}
