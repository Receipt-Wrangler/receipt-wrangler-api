package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildCommentRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	commentRouter := chi.NewRouter()

	commentRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	commentRouter.With(middleware.SetGeneralBodyData("comment", models.Comment{}), middleware.ValidateComment, middleware.SetGroupIdByReceiptId, middleware.ValidateGroupRole(models.VIEWER)).Post("/", handlers.AddComment)
	commentRouter.With(middleware.CanDeleteComment).Delete("/{commentId}", handlers.DeleteComment)

	return commentRouter
}
