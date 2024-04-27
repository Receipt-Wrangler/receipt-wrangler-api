package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildSystemEmailRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	systemEmailRouter := chi.NewRouter()

	systemEmailRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	systemEmailRouter.Post("/", handlers.AddSystemEmail)
	systemEmailRouter.Post("/getSystemEmails", handlers.GetAllSystemEmails)

	return systemEmailRouter
}
