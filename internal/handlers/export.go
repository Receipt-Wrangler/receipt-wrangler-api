package handlers

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm/clause"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

func ExportAllReceiptsFromGroup(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	handler := structs.Handler{
		ErrorMessage: "Error retrieving tags",
		Writer:       w,
		Request:      r,
		ResponseType: constants.TextCsv,
		GroupId:      groupId,
		GroupRole:    models.VIEWER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedRequest := commands.ReceiptPagedRequestCommand{}
			err := pagedRequest.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := pagedRequest.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			token := structs.GetJWT(r)

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipts, _, err := receiptRepository.GetPagedReceiptsByGroupId(token.UserId, groupId, pagedRequest, clause.Associations)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			receiptCsvService := services.NewReceiptCsvService()
			csvBytes, err := receiptCsvService.BuildReceiptCsv(receipts)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Disposition", "attachment; filename=receipts.csv")
			w.WriteHeader(http.StatusOK)
			w.Write(csvBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
