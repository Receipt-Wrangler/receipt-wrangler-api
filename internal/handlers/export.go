package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"

	"github.com/go-chi/chi/v5"
)

func ExportAllReceiptsFromGroup(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	handler := structs.Handler{
		ErrorMessage: "Error exporting receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationZip,
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

			token := structs.GetClaims(r)

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipts, _, err := receiptRepository.
				GetPagedReceiptsByGroupId(
					token.UserId,
					groupId,
					pagedRequest,
					getExportReceiptAssociations(),
				)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			receiptCsvService := services.NewReceiptCsvService()
			zip, err := receiptCsvService.GetZippedCsvFiles(receipts)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
			w.WriteHeader(http.StatusOK)
			w.Write(zip)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func ExportReceiptsById(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	receiptIds := r.Form["receiptIds"]

	handler := structs.Handler{
		ErrorMessage: "Error exporting receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationZip,
		ReceiptIds:   receiptIds,
		GroupRole:    models.VIEWER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			if err != nil {
				return http.StatusInternalServerError, err
			}

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipts, err := receiptRepository.GetReceiptsByIds(receiptIds, getExportReceiptAssociations())
			if err != nil {
				return http.StatusInternalServerError, err
			}

			receiptCsvService := services.NewReceiptCsvService()
			zip, err := receiptCsvService.GetZippedCsvFiles(receipts)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
			w.WriteHeader(http.StatusOK)
			w.Write(zip)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func getExportReceiptAssociations() []string {
	return []string{
		"PaidByUser",
		"ReceiptItems",
		"ReceiptItems.Categories",
		"ReceiptItems.Tags",
		"ReceiptItems.ChargedToUser",
		"ReceiptItems.Receipt",
	}
}
