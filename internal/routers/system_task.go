package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func BuildSystemTaskRouter() *chi.Mux {
	systemTaskRouter := chi.NewRouter()

	systemTaskRouter.Use(middleware.UnifiedAuthMiddleware)
	systemTaskRouter.Post("/getPagedSystemTasks", handlers.GetSystemTasks)
	systemTaskRouter.Post("/getPagedActivities", handlers.GetActivitiesForGroups)
	systemTaskRouter.Post("/rerunActivity/{id}", handlers.RerunActivity)

	return systemTaskRouter
}
