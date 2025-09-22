package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildReceiptImageRouter() *chi.Mux {
	receiptImageRouter := chi.NewRouter()

	receiptImageRouter.Use(middleware.UnifiedAuthMiddleware)
	receiptImageRouter.Get("/{id}", handlers.GetReceiptImage)
	receiptImageRouter.Get("/{id}/download", handlers.DownloadReceiptImage)
	receiptImageRouter.Post("/magicFill", handlers.MagicFillFromImage)
	receiptImageRouter.Delete("/{id}", handlers.RemoveReceiptImage)
	receiptImageRouter.Post("/", handlers.UploadReceiptImage)
	receiptImageRouter.Post("/convertToJpg", handlers.ConvertToJpg)

	return receiptImageRouter
}
