package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildLoginRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	loginRouter := chi.NewRouter()
	loginRouter.With(middleware.SetBodyData, middleware.ValidateLoginData).Post("/", handlers.Login)

	return loginRouter
}
