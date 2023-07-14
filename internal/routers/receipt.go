package routers

import (
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildReceiptRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	receiptRouter := chi.NewRouter()
	receiptRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)

	// swagger:route GET /receipt/{id} Receipt receipt
	//
	// Get receipt
	//
	// This will get a receipt by receipt id[SYSTEM USER]
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
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceipt)
	// receiptRouter.Get("/", handlers.GetReceiptsForGroupIds)

	// swagger:route PUT /receipt/{id} Receipt receipt
	//
	// Update receipt
	//
	// This will update a receipt by receipt id [SYSTEM USER]
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
	receiptRouter.With(middleware.SetReceiptBodyData, middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Put("/{id}", handlers.UpdateReceipt)

	// swagger:route POST /receipt/group/{groupId} Receipt receipt
	//
	// Gets receipts
	//
	// This will return receipts with the option to sort and filter [SYSTEM USER]
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
	receiptRouter.With(middleware.SetGeneralBodyData("pagedRequest", structs.PagedRequest{}), middleware.ValidateGroupRole(models.VIEWER)).Post("/group/{groupId}", handlers.GetPagedReceiptsForGroup)

	// swagger:route POST /receipt/bulkStatusUpdate Receipt receipt
	//
	// Bulk receipt status update
	//
	// This will bulk update receipt statuses with the option of adding a comment to each [SYSTEM USER]
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
	receiptRouter.With(middleware.SetGeneralBodyData("bulkStatusUpdate", structs.BulkStatusUpdate{}), middleware.SetReceiptGroupIds, middleware.BulkValidateGroupRole(models.EDITOR)).Post("/bulkStatusUpdate", handlers.BulkReceiptStatusUpdate)

	// swagger:route POST /receipt/ Receipt receipt
	//
	// Create receipt
	//
	// This will create a receipt [SYSTEM USER]
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
	receiptRouter.With(middleware.SetReceiptBodyData, middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Post("/", handlers.CreateReceipt)

	// swagger:route POST /receipt/{id}/duplicate Receipt receipt
	//
	// Duplicate receipt
	//
	// This will duplicate a receipt [SYSTEM USER]
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
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Post("/{id}/duplicate", handlers.DuplicateReceipt)

	// swagger:route DELETE /receipt/{id} Receipt receipt
	//
	// Delete receipt
	//
	// This will delete a receipt by id [SYSTEM USER]
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
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.DeleteReceipt)
	return receiptRouter
}
