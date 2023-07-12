package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildReceiptImageRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	receiptImageRouter := chi.NewRouter()

	receiptImageRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route GET /receiptImage/{id} ReceiptImage receiptImage
	//
	// Get receipt image
	//
	// This will get a receipt image by id, [SYSTEM USER]
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
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceiptImage)

	// swagger:route DELETE /receiptImage/{id} User user
	//
	// Delete receipt image
	//
	// This will delete a receipt image by id [SYSTEM USER]
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
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.RemoveReceiptImage)

	// swagger:route POST /receiptImage/ User user
	//
	// Uploads a receipt image
	//
	// This will upload a receipt image, [SYSTEM USER]
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
	receiptImageRouter.With(middleware.SetReceiptImageData, middleware.ValidateGroupRole(models.EDITOR)).Post("/", handlers.UploadReceiptImage)

	return receiptImageRouter
}
