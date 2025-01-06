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
)

func GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting system settings",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.APPLICATION_JSON,
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
		ResponseType: constants.APPLICATION_JSON,
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
				updatedSystemSettings.EmailPollingInterval > 0 {
				err = services.StartEmailPolling()
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
