package routers

import (
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildUserRouter() *chi.Mux {
	userRouter := chi.NewRouter()

	// Authenticated routes
	userRouter.With(middleware.UnifiedAuthMiddleware).Get("/", handlers.GetAllUsers)
	userRouter.With(middleware.UnifiedAuthMiddleware, middleware.SetGeneralBodyData("updateProfileCommand", commands.UpdateProfileCommand{})).Put("/updateUserProfile", handlers.UpdateUserProfile)
	userRouter.With(middleware.UnifiedAuthMiddleware, middleware.SetUserData, middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)
	userRouter.With(middleware.UnifiedAuthMiddleware, middleware.SetUserData).Put("/{id}", handlers.UpdateUser)
	userRouter.With(middleware.UnifiedAuthMiddleware, middleware.SetResetPasswordData).Post("/{id}/resetPassword", handlers.ResetPassword)
	userRouter.With(middleware.UnifiedAuthMiddleware, middleware.SetResetPasswordData).Post("/{id}/convertDummyUserToNormalUser", handlers.ConvertDummyUserToNormalUser)
	userRouter.With(middleware.UnifiedAuthMiddleware).Delete("/{id}", handlers.DeleteUser)
	userRouter.With(middleware.UnifiedAuthMiddleware).Delete("/bulk", handlers.BulkDeleteUsers)
	userRouter.With(middleware.UnifiedAuthMiddleware).Get("/amountOwedForUser", handlers.GetAmountOwedForUser)
	userRouter.With(middleware.UnifiedAuthMiddleware).Get("/getUserClaims", handlers.GetClaimsForLoggedInUser)
	userRouter.With(middleware.UnifiedAuthMiddleware).Get("/appData", handlers.GetAppData)

	// Unauthenticated routes
	userRouter.Get("/{username}", handlers.GetUsernameCount)

	return userRouter
}
