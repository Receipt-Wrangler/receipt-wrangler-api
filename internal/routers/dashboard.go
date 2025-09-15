package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildDashboardRouter() *chi.Mux {
	dashboardRouter := chi.NewRouter()

	dashboardRouter.Use(middleware.UnifiedAuthMiddleware)
	dashboardRouter.Get("/{groupId}", handlers.GetDashboardsForUser)
	dashboardRouter.Put("/{dashboardId}", handlers.UpdateDashboard)
	dashboardRouter.Delete("/{dashboardId}", handlers.DeleteDashboard)
	dashboardRouter.Post("/", handlers.CreateDashboard)

	return dashboardRouter
}
