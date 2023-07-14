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

	// swagger:route GET /users/ User user
	//
	// Get users
	//
	// This will get all the users in the system and return a view without sensative information
	//
	//
	//     Produces:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.Get("/", handlers.GetAllUsers)

	// swagger:route GET /users/{username} User user
	//
	// Get username count
	//
	// This will return the number of users in the system with the same username
	//
	//     Consumes:
	//     - text/plain
	//
	//     Produces:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.Get("/{username}", handlers.GetUsernameCount)

	// swagger:route PUT /users/updateUserProfile User user
	//
	// Update user profile
	//
	// This will update the logged in user's user profile
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.SetGeneralBodyData("updateProfileCommand", commands.UpdateProfileCommand{})).Put("/updateUserProfile", handlers.UpdateUserProfile)

	// swagger:route POST /users/ User user
	//
	// Create user
	//
	// This will to create a user, [SYSTEM ADMIN]
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN), middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)

	// swagger:route PUT /users/{id} User user
	//
	// Update user by id
	//
	// This will update a user by id, [SYSTEM ADMIN]
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN)).Put("/{id}", handlers.UpdateUser)

	// swagger:route PUT /users/{id}/resetPassword User user
	//
	// Reset password
	//
	// This will reset a password for a user, [SYSTEM ADMIN]
	//
	//     Consumes:
	//     - application/json
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/resetPassword", handlers.ResetPassword)

	// swagger:route POST /users/{id}/convertDummyUserToNormalUser User user
	//
	// Converts dummy user
	//
	// This will convert a dummy user to a normal system user, [SYSTEM ADMIN]
	//
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}/convertDummyUserToNormaluser", handlers.ConvertDummyUserToNormalUser)

	// swagger:route DELETE /users/{id} User user
	//
	// Delete user
	//
	// This will delete a system user by id [SYSTEM ADMIN]
	//
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.ValidateRole(models.ADMIN)).Delete("/{id}", handlers.DeleteUser)

	// swagger:route GET /users/amountOwedForUser/{groupId} User user
	//
	// Get amount owed for user
	//
	// This will return the amount owed for the logged in user, in the specified group, [SYSTEM USER]
	//
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.With(middleware.ValidateGroupRole(models.VIEWER)).Get("/amountOwedForUser/{groupId}", handlers.GetAmountOwedForUser)

	// swagger:route GET /users/getUserClaims User user
	//
	// Get claims for logged in user
	//
	// This will return the user's token claims for the currently logged in user [SYSTEM USER]
	//
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Security:
	//       api_key:
	//
	//     Responses:
	//       200: Ok
	//       500: Internal Server Error
	userRouter.Get("/getUserClaims", handlers.GetClaimsForLoggedInUser)

	return userRouter
}
