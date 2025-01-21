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

func RerunActivity(w http.ResponseWriter, r *http.Request) {
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)
	inspector, err := wranglerasynq.GetAsynqInspector()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	systemTaskId := chi.URLParam(r, "id")
	systemTaskUintId, err := utils.StringToUint(systemTaskId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	systemTask, err := systemTaskRepository.GetSystemTaskById(systemTaskUintId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if systemTask.Type != models.QUICK_SCAN {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Only quick scan activities can be rerun")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if systemTask.AssociatedSystemTaskId == nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Associated system task id is required to rerun quick scan activity")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parentSystemTask, err := systemTaskRepository.GetSystemTaskById(*systemTask.AssociatedSystemTaskId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if parentSystemTask.AsynqTaskId == "" {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Parent system task does not have an asynq task id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	taskInfo, err := inspector.GetTaskInfo(string(models.QuickScanQueue), parentSystemTask.AsynqTaskId)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var payload wranglerasynq.QuickScanTaskPayload
	err = json.Unmarshal(taskInfo.Payload, &payload)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stringGroupId := utils.UintToString(payload.GroupId)

	handler := structs.Handler{
		ErrorMessage: "Error rerunning activity",
		Writer:       w,
		Request:      r,
		GroupId:      stringGroupId,
		GroupRole:    models.EDITOR,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err = inspector.RunTask(string(models.QuickScanQueue), parentSystemTask.AsynqTaskId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			return 0, nil
		},
	}

	HandleRequest(handler)
}
