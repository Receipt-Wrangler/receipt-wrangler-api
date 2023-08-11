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

	// Authenticated routes
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT).Get("/", handlers.GetAllUsers)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT, middleware.SetGeneralBodyData("updateProfileCommand", commands.UpdateProfileCommand{})).Put("/updateUserProfile", handlers.UpdateUserProfile)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT,middleware.SetUserData, middleware.ValidateRole(models.ADMIN), middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT,middleware.SetUserData, middleware.ValidateRole(models.ADMIN)).Put("/{id}", handlers.UpdateUser)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT,middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/resetPassword", handlers.ResetPassword)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT,middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/convertDummyUserToNormalUser", handlers.ConvertDummyUserToNormalUser)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT, middleware.ValidateRole(models.ADMIN)).Delete("/{id}", handlers.DeleteUser)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT).Get("/amountOwedForUser", handlers.GetAmountOwedForUser)
	userRouter.With(middleware.MoveJWTCookieToHeader,tokenValidator.CheckJWT).Get("/getUserClaims", handlers.GetClaimsForLoggedInUser)
	
	// Unauthenticated routes
	userRouter.Get("/{username}", handlers.GetUsernameCount)

	return userRouter
}
