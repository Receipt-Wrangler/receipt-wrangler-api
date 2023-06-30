package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildNotificationRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	notificationRouter := chi.NewRouter()
	notificationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	notificationRouter.Get("/", handlers.GetNotificationsForUser)

	return notificationRouter
}
