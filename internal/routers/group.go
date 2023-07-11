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

	// swagger:route GET /groups/ Groups group
	//
	// Get groups for user
	//
	// This will get groups for the currently logged in user
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
	groupRouter.Get("/", handlers.GetGroupsForUser)

	// swagger:route GET /groups/{groupId} Groups group
	//
	// Gets a group by Id
	//
	// This will get a group by Id
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
	groupRouter.With(middleware.ValidateGroupRole(models.VIEWER)).Get("/{groupId}", handlers.GetGroupById)

	// swagger:route POST /groups/ Groups group
	//
	// Create group
	//
	// This will create a group
	//
	//     Consumes:
	//     - application/json
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
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{})).Post("/", handlers.CreateGroup)

	// swagger:route PUT /groups/{groupId} Groups group
	//
	// Update a group
	//
	// This will update a group
	//
	//     Consumes:
	//     - application/json
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
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{}), middleware.ValidateGroupRole(models.OWNER)).Put("/{groupId}", handlers.UpdateGroup)

	// swagger:route DELETE /groups/{groupId} Groups group
	//
	// Delete group
	//
	// This will delete a group by id
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
	groupRouter.With(middleware.ValidateGroupRole(models.OWNER), middleware.CanDeleteGroup).Delete("/{groupId}", handlers.DeleteGroup)

	return groupRouter
}
