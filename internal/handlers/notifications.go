package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
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

func GetNotificationCountForUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting notificationCount",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := utils.GetJWT(r)

			result, err := repositories.GetNotificationCountForUser(token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(result)
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
		ResponseType: "",
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

func DeleteNotification(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting notification",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			token := utils.GetJWT(r)

			notification, err := repositories.GetNotificationById(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if notification.UserId != token.UserId {
				return http.StatusForbidden, errors.New("user cannot delete other user's notifications")
			}

			err = repositories.DeleteNotificationById(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
