package routers

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildUserRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	userRouter := chi.NewRouter()

	userRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	userRouter.Get("/", handlers.GetAllUsers)
	userRouter.Get("/{username}", handlers.GetUsernameCount)
	userRouter.With(middleware.SetGeneralBodyData("updateProfileCommand", commands.UpdateProfileCommand{})).Put("/updateUserProfile", handlers.UpdateUserProfile)
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN), middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN)).Put("/{id}", handlers.UpdateUser)
	userRouter.With(middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/resetPassword", handlers.ResetPassword)
	userRouter.With(middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/convertDummyUserToNormalUser", handlers.ConvertDummyUserToNormalUser)
	userRouter.With(middleware.ValidateRole(models.ADMIN)).Delete("/{id}", handlers.DeleteUser)
	userRouter.Get("/amountOwedForUser", handlers.GetAmountOwedForUser)
	userRouter.Get("/getUserClaims", handlers.GetClaimsForLoggedInUser)

	return userRouter
}
