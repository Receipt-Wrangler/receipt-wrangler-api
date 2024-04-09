package routers

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildReceiptRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	receiptRouter := chi.NewRouter()
	receiptRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	receiptRouter.Get("/hasAccess", handlers.HasAccess)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceipt)
	receiptRouter.Put("/{id}", handlers.UpdateReceipt)
	receiptRouter.With(middleware.SetGeneralBodyData("pagedRequest", commands.ReceiptPagedRequestCommand{}), middleware.ValidateGroupRole(models.VIEWER)).Post("/group/{groupId}", handlers.GetPagedReceiptsForGroup)
	receiptRouter.With(middleware.SetGeneralBodyData("BulkStatusUpdateCommand", commands.BulkStatusUpdateCommand{}), middleware.SetReceiptGroupIds, middleware.BulkValidateGroupRole(models.EDITOR)).Post("/bulkStatusUpdate", handlers.BulkReceiptStatusUpdate)
	receiptRouter.Post("/", handlers.CreateReceipt)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Post("/{id}/duplicate", handlers.DuplicateReceipt)
	receiptRouter.Post("/quickScan", handlers.QuickScan)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.DeleteReceipt)
	return receiptRouter
}
