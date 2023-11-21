package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/structs"
)

func CreateDashboard(w http.ResponseWriter, r *http.Request) {

	handler := structs.Handler{
		ErrorMessage: "Error adding dashboard",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {

			return 0, nil
		},
	}

	HandleRequest(handler)
}
