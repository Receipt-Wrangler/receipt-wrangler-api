package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildImportRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	router.Post("/importConfigJson", handlers.ImportConfigJson)

	return router
}
