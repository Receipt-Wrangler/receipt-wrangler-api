package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func GetNotificationsForUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting notifications",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := utils.GetJWT(r)

			notifications, err := repositories.GetNotificationsForUser(token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(notifications)
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

func DeleteAllNotificationsForUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting notifications",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := utils.GetJWT(r)

			err := repositories.DeleteAllNotificationsForUser(token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
