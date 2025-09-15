package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildImportRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.UnifiedAuthMiddleware)
	router.Post("/importConfigJson", handlers.ImportConfigJson)

	return router
}
