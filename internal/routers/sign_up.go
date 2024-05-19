package routers

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildSignUpRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	signUpRouter := chi.NewRouter()

	signUpRouter.Use(middleware.SetGeneralBodyData("signUpCommand", commands.SignUpCommand{}))
	signUpRouter.Post("/", handlers.SignUp)

	return signUpRouter
}
