package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildGroupRouter() *chi.Mux {
	groupRouter := chi.NewRouter()

	groupRouter.Use(middleware.UnifiedAuthMiddleware)
	groupRouter.Get("/", handlers.GetGroupsForUser)
	groupRouter.Get("/{groupId}", handlers.GetGroupById)
	groupRouter.Post("/", handlers.CreateGroup)
	groupRouter.Put("/{groupId}", handlers.UpdateGroup)
	groupRouter.Put("/{groupId}/groupSettings", handlers.UpdateGroupSettings)
	groupRouter.Put("/{groupId}/groupReceiptSettings", handlers.UpdateGroupReceiptSettings)
	groupRouter.With(middleware.CanDeleteGroup).Delete("/{groupId}", handlers.DeleteGroup)
	groupRouter.Post("/{groupId}/pollGroupEmail", handlers.PollGroupEmail)
	groupRouter.Post("/getPagedGroups", handlers.GetPagedGroups)
	groupRouter.Get("/{groupId}/ocrText", handlers.GetOcrTextForGroup)

	return groupRouter
}
