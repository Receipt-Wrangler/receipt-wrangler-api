package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func CreateDashboard(w http.ResponseWriter, r *http.Request) {
	command := commands.UpsertDashboardCommand{}
	command.LoadDataFromRequest(w, r)

	handler := structs.Handler{
		ErrorMessage:  "Error adding dashboard",
		Writer:        w,
		Request:       r,
		GroupId:       command.GroupId,
		GroupRole:     models.VIEWER,
		AllowAllGroup: true,
		ResponseType:  constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			dashboardRepository := repositories.NewDashboardRepository(nil)
			token := structs.GetJWT(r)

			dashboard, err := dashboardRepository.CreateDashboard(command, token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(dashboard)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetDashboardsForUser(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")

	handler := structs.Handler{
		ErrorMessage:  "Error retrieving dashboards",
		Writer:        w,
		Request:       r,
		GroupId:       groupId,
		GroupRole:     models.VIEWER,
		AllowAllGroup: true,
		ResponseType:  constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			dashboardRepository := repositories.NewDashboardRepository(nil)
			token := structs.GetJWT(r)
			uintGroupId, err := simpleutils.StringToUint(groupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			dashboards, err := dashboardRepository.GetDashboardsForUserByGroup(token.UserId, uintGroupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(dashboards)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
