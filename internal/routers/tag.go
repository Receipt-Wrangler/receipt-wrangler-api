package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildTagRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	tagRouter := chi.NewRouter()

	tagRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	tagRouter.Get("/", handlers.GetAllTags)
	tagRouter.Post("/", handlers.CreateTag)
	tagRouter.Post("/getPagedTags", handlers.GetPagedTags)
	tagRouter.Put("/{tagId}", handlers.UpdateTag)
	tagRouter.Delete("/{tagId}", handlers.DeleteTag)
	tagRouter.Get("/{tagName}", handlers.GetTagNameCount)

	return tagRouter
}
