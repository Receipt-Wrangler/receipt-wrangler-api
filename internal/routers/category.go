package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildCategoryRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	categoryRouter := chi.NewRouter()

	categoryRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	categoryRouter.Get("/", handlers.GetAllCategories)
	categoryRouter.Post("/", handlers.CreateCategory)
	categoryRouter.Put("/{categoryId}", handlers.UpdateCategory)
	categoryRouter.Delete("/{categoryId}", handlers.DeleteCategory)
	categoryRouter.Post("/getPagedCategories", handlers.GetPagedCategories)

	return categoryRouter
}
