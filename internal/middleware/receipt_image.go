package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func SetReceiptImageData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Come up with a better way to handdle this
		var fileData models.FileData
		bodyData, err := utils.GetBodyData(w, r)

		if err != nil {
			utils.WriteErrorResponse(w, err, 500)
			return
		}

		marshalErr := json.Unmarshal(bodyData, &fileData)
		if marshalErr != nil {
			utils.WriteErrorResponse(w, marshalErr, 500)
			return
		}

		receipt, err := repositories.GetReceiptById(simpleutils.UintToString(fileData.ReceiptId))
		if err != nil {
			utils.WriteErrorResponse(w, marshalErr, 500)
			return
		}

		ctx := context.WithValue(r.Context(), "fileData", fileData)
		ctx = context.WithValue(ctx, "groupId", simpleutils.UintToString(receipt.GroupId))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetReceiptImageGroupId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Unauthorized access to receipt image."

		receipt, err := services.GetReceiptByReceiptImageId(id)
		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "groupId", simpleutils.UintToString(receipt.GroupId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
