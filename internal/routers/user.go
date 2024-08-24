package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildUserRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	userRouter := chi.NewRouter()

	// Authenticated routes
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Get("/", handlers.GetAllUsers)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT, middleware.SetGeneralBodyData("updateProfileCommand", commands.UpdateProfileCommand{})).Put("/updateUserProfile", handlers.UpdateUserProfile)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT, middleware.SetUserData, middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT, middleware.SetUserData).Put("/{id}", handlers.UpdateUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT, middleware.SetResetPasswordData).Post("/{id}/resetPassword", handlers.ResetPassword)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT, middleware.SetResetPasswordData).Post("/{id}/convertDummyUserToNormalUser", handlers.ConvertDummyUserToNormalUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Delete("/{id}", handlers.DeleteUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Get("/amountOwedForUser", handlers.GetAmountOwedForUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Get("/getUserClaims", handlers.GetClaimsForLoggedInUser)
	userRouter.With(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT).Get("/appData", handlers.GetAppData)

	// Unauthenticated routes
	userRouter.Get("/{username}", handlers.GetUsernameCount)

	return userRouter
}
