package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildCategoryRouter() *chi.Mux {
	categoryRouter := chi.NewRouter()

	categoryRouter.Use(middleware.UnifiedAuthMiddleware)

	categoryRouter.Get("/", handlers.GetAllCategories)
	categoryRouter.Get("/{categoryName}", handlers.GetCategoryNameCount)
	categoryRouter.Post("/", handlers.CreateCategory)
	categoryRouter.Put("/{categoryId}", handlers.UpdateCategory)
	categoryRouter.Delete("/{categoryId}", handlers.DeleteCategory)
	categoryRouter.Post("/getPagedCategories", handlers.GetPagedCategories)

	return categoryRouter
}
