package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildGroupRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	groupRouter := chi.NewRouter()

	groupRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	groupRouter.Get("/", handlers.GetGroupsForUser)
	groupRouter.Get("/{groupId}", handlers.GetGroupById)
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{})).Post("/", handlers.CreateGroup)
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{})).Put("/{groupId}", handlers.UpdateGroup)
	groupRouter.Put("/{groupId}/groupSettings", handlers.UpdateGroupSettings)
	groupRouter.With(middleware.CanDeleteGroup).Delete("/{groupId}", handlers.DeleteGroup)
	groupRouter.Post("/{groupId}/pollGroupEmail", handlers.PollGroupEmail)

	return groupRouter
}
