package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildCommentRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	commentRouter := chi.NewRouter()

	commentRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route POST /comment/ Comment comment
	//
	// Add comment
	//
	// This will add a comment to a receipt, [SYSTEM USER]
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
	commentRouter.With(middleware.SetGeneralBodyData("comment", models.Comment{}), middleware.ValidateComment, middleware.SetGroupIdByReceiptId, middleware.ValidateGroupRole(models.VIEWER)).Post("/", handlers.AddComment)

	// swagger:route DELETE /comment/{commentId} Comment comment
	//
	// Delete comment
	//
	// This will delete a comment by id [SYSTEM User]
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
	commentRouter.With(middleware.CanDeleteComment).Delete("/{commentId}", handlers.DeleteComment)

	return commentRouter
}
