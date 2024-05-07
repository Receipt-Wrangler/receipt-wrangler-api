package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildPromptRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	router.Post("/getPagedPrompts", handlers.GetPagedPrompts)

	return router
}
