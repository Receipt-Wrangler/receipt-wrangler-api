package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildDataRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	dashboardRouter := chi.NewRouter()

	dashboardRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	dashboardRouter.Get("/{groupId}/ocrText", handlers.GetOcrTextForGroup)

	return dashboardRouter
}
