package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func Search(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error searching",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			// Get groups that user has access to
			// Search receipts and receipt items on name that matches string
			db := db.GetDB()
			var receipts []models.Receipt
			//var items []models.Item

			results := make([]structs.SearchResult, 0)

			token := utils.GetJWT(r)
			groupIds, err := repositories.GetGroupIdsByUserId(utils.UintToString(token.UserId))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Table("receipts").Where("group_id IN ? AND name LIKE ?", groupIds, "test").Find(&receipts).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			for _, receipt := range receipts {
				results = append(results, structs.SearchResult{
					ID:   receipt.ID,
					Name: receipt.Name,
					Type: "Receipt",
				})
			}

			bytes, err := utils.MarshalResponseData(results)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
