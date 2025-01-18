package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"receipt-wrangler/api/internal/wranglerasynq"
)

func GetSystemTasks(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting system tasks",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.GetSystemTaskCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			systemTaskRepository := repositories.NewSystemTaskRepository(nil)
			systemTasks, count, err := systemTaskRepository.GetPagedSystemTasks(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			pagedData := structs.PagedData{}
			data := make([]any, 0)

			for i := 0; i < len(systemTasks); i++ {
				data = append(data, systemTasks[i])
			}

			pagedData.Data = data
			pagedData.TotalCount = count

			responseBytes, err := utils.MarshalResponseData(pagedData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(responseBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetActivitiesForGroups(w http.ResponseWriter, r *http.Request) {
	command := commands.PagedActivityRequestCommand{}
	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stringGroupIds := make([]string, 0)
	for _, groupId := range command.GroupIds {
		stringGroupIds = append(stringGroupIds, utils.UintToString(groupId))
	}

	handler := structs.Handler{
		ErrorMessage: "Error getting group activities",
		Writer:       w,
		Request:      r,
		GroupIds:     stringGroupIds,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {

			vErr := command.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			systemTaskRepository := repositories.NewSystemTaskRepository(nil)
			activities, count, err := systemTaskRepository.GetPagedActivities(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = wranglerasynq.SetActivityCanBeRestarted(&activities)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			pagedData := structs.PagedData{}
			data := make([]any, 0)

			for i := 0; i < len(activities); i++ {
				data = append(data, activities[i])
			}

			pagedData.Data = data
			pagedData.TotalCount = count

			responseBytes, err := utils.MarshalResponseData(pagedData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(responseBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
