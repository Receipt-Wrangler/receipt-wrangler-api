package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func SetReceiptImageData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// TODO: Come up with a better way to handdle this
			var receipt models.FileData
			bodyData, err := utils.GetBodyData(w, r)

			if err != nil {
				utils.WriteErrorResponse(w, err, 500)
				return
			}

			marshalErr := json.Unmarshal(bodyData, &receipt)
			if marshalErr != nil {
				utils.WriteErrorResponse(w, marshalErr, 500)
				return
			}
			ctx := context.WithValue(r.Context(), "fileData", receipt)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func ValidateReceiptImageAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Unauthorized access to receipt image."

		if len(id) > 0 {
			token := utils.GetJWT(r)

			receipt, err := services.GetReceiptByReceiptImageId(id)
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(w, errMsg, 500)
				return
			}

			hasAccess, err := services.UserHasAccessToReceipt(token.UserId, strconv.FormatUint(uint64(receipt.ID), 10))
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(w, errMsg, 500)
				return
			}

			if !hasAccess {
				utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
		return
	})
}
