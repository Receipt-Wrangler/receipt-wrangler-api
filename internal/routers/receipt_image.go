package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildReceiptImageRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	receiptImageRouter := chi.NewRouter()

	receiptImageRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	receiptImageRouter.Get("/{id}", handlers.GetReceiptImage)
	receiptImageRouter.Post("/magicFill", handlers.MagicFillFromImage)
	receiptImageRouter.Delete("/{id}", handlers.RemoveReceiptImage)
	receiptImageRouter.Post("/", handlers.UploadReceiptImage)
	receiptImageRouter.Post("/convertToJpg", handlers.ConvertToJpg)

	return receiptImageRouter
}
