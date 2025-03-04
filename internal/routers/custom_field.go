package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildCustomFieldRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	router.Post("/", handlers.GetPagedCustomFields)

	return router
}
