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
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceipt)
	receiptRouter.With(middleware.SetReceiptBodyData, middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Put("/{id}", handlers.UpdateReceipt)
	receiptRouter.With(middleware.SetGeneralBodyData("pagedRequest", commands.PagedRequestCommand{}), middleware.ValidateGroupRole(models.VIEWER)).Post("/group/{groupId}", handlers.GetPagedReceiptsForGroup)
	receiptRouter.With(middleware.SetGeneralBodyData("BulkStatusUpdateCommand", commands.BulkStatusUpdateCommand{}), middleware.SetReceiptGroupIds, middleware.BulkValidateGroupRole(models.EDITOR)).Post("/bulkStatusUpdate", handlers.BulkReceiptStatusUpdate)
	receiptRouter.With(middleware.SetReceiptBodyData, middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Post("/", handlers.CreateReceipt)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Post("/{id}/duplicate", handlers.DuplicateReceipt)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.DeleteReceipt)
	return receiptRouter
}
