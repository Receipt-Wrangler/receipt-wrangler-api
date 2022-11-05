package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

func SetReceiptBodyData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			// TODO: Come up with a better way to handdle this
			var receipt models.Receipt
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
			ctx := context.WithValue(r.Context(), "receipt", receipt)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func ValidateReceiptAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Unauthorized access to receipt image."

		if len(id) > 0 {
			db := db.GetDB()
			token := utils.GetJWT(r)
			var receipt models.Receipt

			err := db.Model(models.Receipt{}).Where("id = ?", id).Select("owned_by_user_id").Find(&receipt).Error
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(w, errMsg, 500)
				return
			}

			if receipt.OwnedByUserID != token.UserId {
				middleware_logger.Print(errMsg)
				utils.WriteCustomErrorResponse(w, errMsg, 403)
				return
			}
		}

		next.ServeHTTP(w, r)
		return
	})
}

func ValidateReceipt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := structs.ValidatorError{
			Errors: make(map[string]string),
		}
		receipt := r.Context().Value("receipt").(models.Receipt)

		requiredNameMsg := "Name is required"
		requiredAmountMsg := "Amount must be greater than zero"

		if len(receipt.Name) == 0 {
			err.Errors["name"] = requiredNameMsg
		}

		if receipt.Amount.LessThanOrEqual(decimal.Zero) {
			err.Errors["amount"] = requiredAmountMsg
		}

		if receipt.Date.IsZero() {
			err.Errors["date"] = "Date is required"
		}

		for i, item := range receipt.ReceiptItems {
			basePath := fmt.Sprintf("receiptItems.%s", fmt.Sprint(i))

			if len(item.Name) == 0 {
				err.Errors[basePath+".name"] = requiredNameMsg
			}

			if decimal.Zero.Equal(item.Amount) {
				err.Errors[basePath+"amount"] = requiredAmountMsg
			}

			if item.Amount.GreaterThan(receipt.Amount) {
				err.Errors[basePath+"amount"] = "Value cannot be larger than total receipt amount"
			}
		}

		if len(err.Errors) > 0 {
			middleware_logger.Print(err.Errors, r)
			utils.WriteValidatorErrorResponse(w, err, 400)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}
