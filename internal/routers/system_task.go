package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildSystemTaskRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	systemTaskRouter := chi.NewRouter()

	systemTaskRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	systemTaskRouter.Post("/getPagedSystemTasks", handlers.GetSystemTasks)
	systemTaskRouter.Post("/getPagedActivities", handlers.GetActivitiesForGroups)

	return systemTaskRouter
}
