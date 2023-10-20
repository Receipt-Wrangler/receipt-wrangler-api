package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildReceiptImageRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	receiptImageRouter := chi.NewRouter()

	receiptImageRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceiptImage)
	receiptImageRouter.Post("/magicFill", handlers.MagicFillFromImage)
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.RemoveReceiptImage)
	receiptImageRouter.Post("/", handlers.UploadReceiptImage)

	return receiptImageRouter
}
