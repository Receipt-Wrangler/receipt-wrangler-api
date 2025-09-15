package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildSystemEmailRouter() *chi.Mux {
	systemEmailRouter := chi.NewRouter()

	systemEmailRouter.Use(middleware.UnifiedAuthMiddleware)
	systemEmailRouter.Get("/{id}", handlers.GetSystemEmailById)
	systemEmailRouter.Put("/{id}", handlers.UpdateSystemEmail)
	systemEmailRouter.Delete("/{id}", handlers.DeleteSystemEmail)
	systemEmailRouter.Post("/", handlers.AddSystemEmail)
	systemEmailRouter.Post("/checkConnectivity", handlers.CheckConnectivity)
	systemEmailRouter.Post("/getSystemEmails", handlers.GetAllSystemEmails)

	return systemEmailRouter
}
