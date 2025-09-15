package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildSearchRouter() *chi.Mux {
	searchRouter := chi.NewRouter()

	searchRouter.Use(middleware.UnifiedAuthMiddleware)
	searchRouter.Get("/", handlers.Search)

	return searchRouter
}
