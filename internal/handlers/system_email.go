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

func GetAllSystemEmails(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving system emails",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
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

func AddSystemEmail(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error adding system email",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertSystemEmailCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErr := command.Validate()
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return http.StatusBadRequest, nil
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
