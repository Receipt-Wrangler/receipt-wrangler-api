package middleware

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func SetReceiptGroupId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		errMsg := "Error accessing receipt."

		receiptRepository := repositories.NewReceiptRepository(nil)
		groupId, err := receiptRepository.GetReceiptGroupIdByReceiptId(id)
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
		db := repositories.GetDB()
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
