package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildDashboardRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	dashboardRouter := chi.NewRouter()

	dashboardRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	dashboardRouter.Get("/{groupId}", handlers.GetDashboardsForUser)
	dashboardRouter.Put("/{dashboardId}", handlers.UpdateDashboard)
	dashboardRouter.Post("/", handlers.CreateDashboard)

	return dashboardRouter
}
