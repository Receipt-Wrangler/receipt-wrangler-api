package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
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
				middleware_logger.Print(err.Error())
				utils.WriteErrorResponse(w, err, 500)
				return
			}

			err = json.Unmarshal(bodyData, &receipt)
			if err != nil {
				middleware_logger.Print(err.Error())
				utils.WriteErrorResponse(w, err, 500)
				return
			}
			ctx := context.WithValue(r.Context(), "receipt", receipt)
			ctx = context.WithValue(ctx, "groupId", simpleutils.UintToString(receipt.GroupId))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func SetReceiptGroupId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Error accessing receipt."

		groupId, err := repositories.GetReceiptGroupIdByReceiptId(id)
		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "groupId", simpleutils.UintToString(groupId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetReceiptGroupIds(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := db.GetDB()
		receiptIds := r.Context().Value("receiptIds").([]uint)
		groupIds := make([]string, len(receiptIds))
		errMsg := "Error accessing receipt."
		var receipts []models.Receipt

		err := db.Model(models.Receipt{}).Where("id IN ?", receiptIds).Select("group_id").Find(&receipts).Error
		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, http.StatusForbidden)
			return
		}

		for i := 0; i < len(receiptIds); i++ {
			groupIds[i] = simpleutils.UintToString(receipts[i].GroupId)
		}

		ctx := context.WithValue(r.Context(), "groupIds", groupIds)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ValidateReceipt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := utils.GetJWT(r)
		err := structs.ValidatorError{
			Errors: make(map[string]string),
		}
		receipt := r.Context().Value("receipt").(models.Receipt)

		requiredNameMsg := "Name is required"
		requiredAmountMsg := "Amount must be greater than zero"
		requiredStatusMsg := "Status is required"

		if len(receipt.Name) == 0 {
			err.Errors["name"] = requiredNameMsg
		}

		if receipt.Amount.LessThanOrEqual(decimal.Zero) {
			err.Errors["amount"] = requiredAmountMsg
		}

		if len(receipt.Status) == 0 {
			err.Errors["status"] = requiredStatusMsg
		}

		if !utils.Contains(constants.ReceiptStatuses(), receipt.Status) {
			err.Errors["status"] = "Invalid status"
		}

		if receipt.Date.IsZero() {
			err.Errors["date"] = "Date is required"
		}

		// Validate that users aren't commenting as other users on create by manipulating the body data
		if r.Method == "POST" {
			for i, comment := range receipt.Comments {
				basePath := fmt.Sprintf("comments.%s", fmt.Sprint(i))

				if *comment.UserId != token.UserId {
					err.Errors[basePath+".userId"] = "User cannot comment as anyone other than themselves."
				}
			}
		}

		// Validate the receipt data
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
