package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
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
	errorMsg := "Error getting group activities"
	command := commands.PagedActivityRequestCommand{}
	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	stringGroupIds := make([]string, 0)
	for _, groupId := range command.GroupIds {
		stringGroupIds = append(stringGroupIds, utils.UintToString(groupId))
	}

	handler := structs.Handler{
		ErrorMessage: errorMsg,
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

func RerunActivity(w http.ResponseWriter, r *http.Request) {
	errorMsg := "Error rerunning activity"
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)
	inspector, err := wranglerasynq.GetAsynqInspector()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}
	defer inspector.Close()

	systemTaskId := chi.URLParam(r, "id")
	systemTaskUintId, err := utils.StringToUint(systemTaskId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	systemTask, err := systemTaskRepository.GetSystemTaskById(systemTaskUintId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	if systemTask.Type != models.QUICK_SCAN && systemTask.Type != models.EMAIL_UPLOAD {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Only quick scan and email upload activities can be rerun")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	queueName, err := wranglerasynq.SystemTaskToQueueName(systemTask.Type)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	taskInfo, err := inspector.GetTaskInfo(queueName, systemTask.AsynqTaskId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	var payload wranglerasynq.RerunTaskPayload
	err = json.Unmarshal(taskInfo.Payload, &payload)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
		return
	}

	stringGroupId := utils.UintToString(payload.GroupId)
	if payload.GroupSettingsId > 0 {
		groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)

		stringId := utils.UintToString(payload.GroupSettingsId)
		groupSettings, err := groupSettingsRepository.GetGroupSettingsById(stringId)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			utils.WriteCustomErrorResponse(w, errorMsg, http.StatusInternalServerError)
			return
		}

		stringGroupId = utils.UintToString(groupSettings.GroupId)
	}

	handler := structs.Handler{
		ErrorMessage: errorMsg,
		Writer:       w,
		Request:      r,
		GroupId:      stringGroupId,
		GroupRole:    models.EDITOR,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err = inspector.RunTask(queueName, systemTask.AsynqTaskId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			return 0, nil
		},
	}

	HandleRequest(handler)
}
