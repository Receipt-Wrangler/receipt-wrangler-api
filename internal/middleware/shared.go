package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func SetGeneralBodyData(contextKey string, dataType interface{}) (mw func(http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyData, err := utils.GetBodyData(w, r)

			if err != nil {
				utils.WriteErrorResponse(w, err, 500)
				return
			}

			switch dataType.(type) {
			case models.Group:
				var group models.Group
				err = json.Unmarshal(bodyData, &group)
				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, group)
				serveWithContext(r, w, h, ctx)

			case models.Comment:
				var comment models.Comment
				err = json.Unmarshal(bodyData, &comment)
				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), "receiptId", simpleutils.UintToString(comment.ReceiptId))
				ctx = context.WithValue(ctx, contextKey, comment)
				serveWithContext(r, w, h, ctx)

			case commands.BulkStatusUpdateCommand:
				var bulkResolve commands.BulkStatusUpdateCommand
				err = json.Unmarshal(bodyData, &bulkResolve)

				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, bulkResolve)
				ctx = context.WithValue(ctx, "receiptIds", bulkResolve.ReceiptIds)
				serveWithContext(r, w, h, ctx)

			case structs.PagedRequest:
				var pagedRequest structs.PagedRequest
				err = json.Unmarshal(bodyData, &pagedRequest)

				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, pagedRequest)
				serveWithContext(r, w, h, ctx)

			case commands.UpdateProfileCommand:
				var command commands.UpdateProfileCommand
				err = json.Unmarshal(bodyData, &command)

				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, command)
				serveWithContext(r, w, h, ctx)

			case commands.SignUpCommand:
				var command commands.SignUpCommand
				err = json.Unmarshal(bodyData, &command)

				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}

				ctx := context.WithValue(r.Context(), contextKey, command)
				serveWithContext(r, w, h, ctx)

			default:
				return
			}
		})
	}
	return
}

func SetGroupIdByReceiptId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var receipt models.Receipt
		db := db.GetDB()

		receiptId := r.Context().Value("receiptId").(string)
		if len(receiptId) == 0 {
			middleware_logger.Println("Invalid context receiptId", r)
			utils.WriteCustomErrorResponse(w, "Malformed request", http.StatusBadRequest)
			return
		}

		err := db.Model(&models.Receipt{}).Where("id = ?", receiptId).Select("group_id").Find(&receipt).Error
		if err != nil {
			middleware_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, "Error getting receipt", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "groupId", simpleutils.UintToString(receipt.GroupId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkError(err error, w http.ResponseWriter) bool {
	if err != nil {
		middleware_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return true
	}
	return false
}

func serveWithContext(r *http.Request, w http.ResponseWriter, h http.Handler, ctx context.Context) {
	h.ServeHTTP(w, r.WithContext(ctx))
}
