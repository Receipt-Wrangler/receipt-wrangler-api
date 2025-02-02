package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"receipt-wrangler/api/internal/wranglerasynq"
)

func GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting system settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
			systemSettings, err := systemSettingsRepository.GetSystemSettings()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := utils.MarshalResponseData(systemSettings)
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

func UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating system settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
			command := commands.UpsertSystemSettingsCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			previousSystemSettings, err := systemSettingsRepository.GetSystemSettings()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			updatedSystemSettings, err := systemSettingsRepository.UpdateSystemSettings(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if previousSystemSettings.EmailPollingInterval != updatedSystemSettings.EmailPollingInterval &&
				updatedSystemSettings.EmailPollingInterval > 0 && config.GetDeployEnv() != "test" {
				err = wranglerasynq.StartEmailPolling()
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			responseBytes, err := utils.MarshalResponseData(updatedSystemSettings)
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

func RestartTaskServer(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error restarting task server, please restart the entire application",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err := wranglerasynq.RestartEmbeddedAsynqServer()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}
	HandleRequest(handler)
}
