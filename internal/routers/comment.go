package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildCommentRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	commentRouter := chi.NewRouter()

	commentRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	commentRouter.Post("/", handlers.AddComment)
	commentRouter.Delete("/{commentId}", handlers.DeleteComment)

	return commentRouter
}
