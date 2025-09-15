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

func GetAllSystemEmails(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving system emails",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedData := structs.PagedData{
				Data:       []any{},
				TotalCount: 0,
			}
			pagedCommand := commands.PagedRequestCommand{}
			err := pagedCommand.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := pagedCommand.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return http.StatusBadRequest, nil
			}

			systemEmailRepository := repositories.NewSystemEmailRepository(nil)
			systemEmails, err := systemEmailRepository.GetPagedSystemEmails(pagedCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			count, err := systemEmailRepository.GetCount("system_emails", "")
			if err != nil {
				return http.StatusInternalServerError, err
			}

			for i := 0; i < len(systemEmails); i++ {
				pagedData.Data = append(pagedData.Data, systemEmails[i])
			}
			pagedData.TotalCount = count

			bytes, err := utils.MarshalResponseData(pagedData)
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

func GetSystemEmailById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving system emails",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			systemEmailId := chi.URLParam(r, "id")
			systemEmailRepository := repositories.NewSystemEmailRepository(nil)

			systemEmail, err := systemEmailRepository.GetSystemEmailById(systemEmailId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(systemEmail)
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

func AddSystemEmail(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error adding system email",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertSystemEmailCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate(true)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return http.StatusBadRequest, nil
			}

			systemEmailRepository := repositories.NewSystemEmailRepository(nil)
			systemEmail, err := systemEmailRepository.AddSystemEmail(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(systemEmail)
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

func UpdateSystemEmail(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating system email",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertSystemEmailCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate(false)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return http.StatusBadRequest, nil
			}

			id := chi.URLParam(r, "id")
			updatePassword := false
			if r.URL.Query().Get("updatePassword") == "true" {
				updatePassword = true
			}

			systemEmailRepository := repositories.NewSystemEmailRepository(nil)
			systemEmail, err := systemEmailRepository.UpdateSystemEmail(id, command, updatePassword)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(systemEmail)
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

func DeleteSystemEmail(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting system email",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			systemEmailRepository := repositories.NewSystemEmailRepository(nil)

			err := systemEmailRepository.DeleteSystemEmail(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func CheckConnectivity(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Could not connect with credentials",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.CheckEmailConnectivityCommand{}
			token := structs.GetClaims(r)

			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return http.StatusBadRequest, nil
			}

			systemEmailService := services.NewSystemEmailService(nil)
			systemTask, err := systemEmailService.CheckEmailConnectivity(command, token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(systemTask)
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
