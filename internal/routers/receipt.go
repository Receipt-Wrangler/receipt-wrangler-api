package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildReceiptRouter() *chi.Mux {
	receiptRouter := chi.NewRouter()
	receiptRouter.Use(middleware.UnifiedAuthMiddleware)
	receiptRouter.Get("/hasAccess", handlers.HasAccess)
	receiptRouter.Get("/{id}", handlers.GetReceipt)
	receiptRouter.Put("/{id}", handlers.UpdateReceipt)
	receiptRouter.Post("/group/{groupId}", handlers.GetPagedReceiptsForGroup)
	receiptRouter.Post("/bulkStatusUpdate", handlers.BulkReceiptStatusUpdate)
	receiptRouter.Post("/", handlers.CreateReceipt)
	receiptRouter.Post("/{id}/duplicate", handlers.DuplicateReceipt)
	receiptRouter.Post("/quickScan", handlers.QuickScan)
	receiptRouter.Delete("/{id}", handlers.DeleteReceipt)
	return receiptRouter
}
