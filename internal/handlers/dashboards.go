package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func CreateDashboard(w http.ResponseWriter, r *http.Request) {
	command := commands.UpsertDashboardCommand{}
	command.LoadDataFromRequest(w, r)

	handler := structs.Handler{
		ErrorMessage: "Error adding dashboard",
		Writer:       w,
		Request:      r,
		GroupId:      command.GroupId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.APPLICATION_JSON,
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
