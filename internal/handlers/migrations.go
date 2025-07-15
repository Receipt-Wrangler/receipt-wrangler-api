package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

func MigrateItemsToShares(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error migrating items to shares.",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err := services.MigrateItemsToShares()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			return 0, nil
		},
	}
	HandleRequest(handler)
}
