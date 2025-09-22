package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildApiKeyRouter() *chi.Mux {
	apiKeyRouter := chi.NewRouter()

	// Authenticated routes
	apiKeyRouter.With(middleware.UnifiedAuthMiddleware).Post("/", handlers.CreateApiKey)
	apiKeyRouter.With(middleware.UnifiedAuthMiddleware).Post("/paged", handlers.GetPagedApiKeys)
	apiKeyRouter.With(middleware.UnifiedAuthMiddleware).Put("/{id}", handlers.UpdateApiKey)
	apiKeyRouter.With(middleware.UnifiedAuthMiddleware).Delete("/{id}", handlers.DeleteApiKey)

	return apiKeyRouter
}
