package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func CreateDashboard(w http.ResponseWriter, r *http.Request) {
	command := commands.UpsertDashboardCommand{}
	vErr, err := command.LoadDataFromRequestAndValidate(w, r)

	handler := structs.Handler{
		ErrorMessage: "Error adding dashboard",
		Writer:       w,
		Request:      r,
		GroupId:      command.GroupId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			dashboardRepository := repositories.NewDashboardRepository(nil)
			token := structs.GetClaims(r)

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
		ErrorMessage: "Error retrieving dashboards",
		Writer:       w,
		Request:      r,
		GroupId:      groupId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			dashboardRepository := repositories.NewDashboardRepository(nil)
			token := structs.GetClaims(r)
			uintGroupId, err := utils.StringToUint(groupId)
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

func UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardId := chi.URLParam(r, "dashboardId")
	dashboardRepository := repositories.NewDashboardRepository(nil)
	uintDashboardId, err := utils.StringToUint(dashboardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dashboard, err := dashboardRepository.GetDashboardById(uintDashboardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stringGroupId := utils.UintToString(dashboard.GroupID)

	handler := structs.Handler{
		ErrorMessage: "Error updating dashboard",
		Writer:       w,
		Request:      r,
		GroupId:      stringGroupId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			if dashboard.UserID != token.UserId {
				return http.StatusForbidden, nil
			}

			command := commands.UpsertDashboardCommand{}
			vErr, err := command.LoadDataFromRequestAndValidate(w, r)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			if err != nil {
				return http.StatusInternalServerError, err
			}

			dashboard, err := dashboardRepository.UpdateDashboardById(uintDashboardId, command)
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

func DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardId := chi.URLParam(r, "dashboardId")
	uintDashboardId, err := utils.StringToUint(dashboardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dashboardRepository := repositories.NewDashboardRepository(nil)
	dashboard, err := dashboardRepository.GetDashboardById(uintDashboardId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stringDashboardId := utils.UintToString(dashboard.GroupID)

	handler := structs.Handler{
		ErrorMessage: "Error deleteing dashboard",
		Writer:       w,
		Request:      r,
		GroupId:      stringDashboardId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			if dashboard.UserID != token.UserId {
				return http.StatusForbidden, nil
			}

			dashboardRepository := repositories.NewDashboardRepository(nil)
			err := dashboardRepository.DeleteDashboardById(uintDashboardId)
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
