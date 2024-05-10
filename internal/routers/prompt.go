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
	router.Get("/{id}", handlers.GetPromptById)
	router.Put("/{id}", handlers.UpdatePromptById)
	router.Delete("/{id}", handlers.DeletePromptById)
	router.Post("/", handlers.CreatePrompt)
	router.Post("/getPagedPrompts", handlers.GetPagedPrompts)

	return router
}
