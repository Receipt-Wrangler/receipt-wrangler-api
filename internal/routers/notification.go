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

	// swagger:route GET /api/notifications Notifications listNotifications
	//
	// List notifications for logged in user
	//
	// This will get all the notifications for the currently logged in user
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//
	//
	//     Responses:
	//       default: genericError
	//       200: Ok
	//       500: Internal Server Error
	notificationRouter.Get("/", handlers.GetNotificationsForUser)
	notificationRouter.Get("/notificationCount", handlers.GetNotificationCountForUser)
	notificationRouter.Delete("/", handlers.DeleteAllNotificationsForUser)
	notificationRouter.Delete("/{id}", handlers.DeleteNotification)

	return notificationRouter
}
