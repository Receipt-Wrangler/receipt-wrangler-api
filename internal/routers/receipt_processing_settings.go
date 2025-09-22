package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildReceiptProcessingSettingsRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.UnifiedAuthMiddleware)
	router.Get("/{id}", handlers.GetReceiptProcessingSettingsById)
	router.Post("/", handlers.CreateReceiptProcessingSettings)
	router.Post("/getPagedProcessingSettings", handlers.GetPagedReceiptProcessingSettings)
	router.Post("/checkConnectivity", handlers.CheckReceiptProcessingSettingsConnectivity)
	router.Put("/{id}", handlers.UpdateReceiptProcessingSettingsById)
	router.Delete("/{id}", handlers.DeleteReceiptProcessingSettingsById)

	return router
}
