package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildCustomFieldRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.UnifiedAuthMiddleware)

	router.Get("/{id}", handlers.GetCustomFieldById)
	router.Delete("/{id}", handlers.DeleteCustomField)
	router.Post("/getPagedCustomFields", handlers.GetPagedCustomFields)
	router.Post("/", handlers.CreateCustomField)

	return router
}
