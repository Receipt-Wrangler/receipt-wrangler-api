package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildMigrationRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	migrationRouter := chi.NewRouter()

	migrationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	migrationRouter.Post("/items-to-shares", handlers.MigrateItemsToShares)

	return migrationRouter
}
