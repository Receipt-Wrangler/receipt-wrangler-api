package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildExportRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	exportRouter := chi.NewRouter()

	exportRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	exportRouter.Post("/", handlers.ExportReceiptsById)
	exportRouter.Post("/{groupId}", handlers.ExportAllReceiptsFromGroup)

	return exportRouter
}
