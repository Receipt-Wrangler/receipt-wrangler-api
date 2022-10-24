package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

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

func ValidateAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Unauthorized access to receipt image."

		if len(id) > 0 {
			db := db.GetDB()
			token := utils.GetJWT(r)
			var receipt models.Receipt
			var fileData models.FileData

			err := db.Model(models.FileData{}).Where("id = ?", id).Select("receipt_id").First(&fileData).Error
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(w, errMsg, 500)
				return
			}

			err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("owned_by_user_id").Find(&receipt).Error
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(w, errMsg, 500)
				return
			}

			if receipt.OwnedByUserID != token.UserId {
				middleware_logger.Print("Unauthorized access")
				utils.WriteCustomErrorResponse(w, errMsg, 403)
				return
			}
		}

		next.ServeHTTP(w, r)
		return
	})
}
