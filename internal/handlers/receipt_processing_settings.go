package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetPagedReceiptProcessingSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting receipt processing settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.PagedRequestCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
			receiptProcessingSettings, count, err := receiptProcessingSettingsRepository.GetPagedReceiptProcessingSettings(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			pagedData := structs.PagedData{}
			data := make([]interface{}, 0)
			for i := 0; i < len(receiptProcessingSettings); i++ {
				data = append(data, receiptProcessingSettings[i])
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

func CreateReceiptProcessingSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating receipt processing settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertReceiptProcessingSettingsCommand{}

			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate(false)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
			settings, err := receiptProcessingSettingsRepository.CreateReceiptProcessingSettings(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := utils.MarshalResponseData(settings)
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

func GetReceiptProcessingSettingsById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting receipt processing settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")

			receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
			settings, err := receiptProcessingSettingsRepository.GetReceiptProcessingSettingsById(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := utils.MarshalResponseData(settings)
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

func UpdateReceiptProcessingSettingsById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating receipt processing settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			updateKey := r.URL.Query().Get("updateKey") == "true"

			command := commands.UpsertReceiptProcessingSettingsCommand{}

			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate(updateKey)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}

			receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
			settings, err := receiptProcessingSettingsRepository.UpdateReceiptProcessingSettingsById(id, updateKey, command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := utils.MarshalResponseData(settings)
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

func DeleteReceiptProcessingSettingsById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting receipt processing settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")

			receiptProcessingSettingsRepository := repositories.NewReceiptProcessingSettings(nil)
			err := receiptProcessingSettingsRepository.DeleteReceiptProcessingSettingsById(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func CheckReceiptProcessingSettingsConnectivity(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error checking connectivity",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			command := commands.CheckReceiptProcessingSettingsCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}
			var aiClient *services.AiService
			var decryptKey bool

			if command.ID > 0 && command.UpsertReceiptProcessingSettingsCommand.IsEmpty() {
				stringId := utils.UintToString(command.ID)

				client, clientErr := services.NewAiService(stringId)
				if clientErr != nil {
					return http.StatusInternalServerError, clientErr
				}

				aiClient = client
				decryptKey = true
			} else {
				receiptProcessingSettings := models.ReceiptProcessingSettings{
					BaseModel:   models.BaseModel{ID: command.ID},
					Name:        command.Name,
					Description: command.Description,
					AiType:      command.AiType,
					Url:         command.Url,
					Key:         command.Key,
					Model:       command.Model,
					OcrEngine:   &command.OcrEngine,
					PromptId:    command.PromptId,
				}

				client := services.AiService{}
				client.ReceiptProcessingSettings = receiptProcessingSettings

				aiClient = &client
				decryptKey = false
			}

			systemTask, err := aiClient.CheckConnectivity(token.UserId, decryptKey)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := utils.MarshalResponseData(systemTask)
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
