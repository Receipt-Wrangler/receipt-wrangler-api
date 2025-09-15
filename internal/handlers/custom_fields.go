package handlers

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetPagedCustomFields(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting custom fields",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.USER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedData := structs.PagedData{}
			pagedRequestCommand := commands.PagedRequestCommand{}
			err := pagedRequestCommand.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := pagedRequestCommand.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			customFieldsRepository := repositories.NewCustomFieldRepository(nil)
			customFields, count, err := customFieldsRepository.GetPagedCustomFields(pagedRequestCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			anyData := make([]any, len(customFields))
			for i := 0; i < len(customFields); i++ {
				anyData[i] = customFields[i]
			}

			pagedData.Data = anyData
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

func CreateCustomField(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating custom field",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.USER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertCustomFieldCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			token := structs.GetClaims(r)
			customFieldsRepository := repositories.NewCustomFieldRepository(nil)
			customField, err := customFieldsRepository.CreateCustomField(command, &token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			bytes, err := json.Marshal(customField)
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

func GetCustomFieldById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting custom field",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.USER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			customFieldId := chi.URLParam(r, "id")
			customFieldIdUint, err := utils.StringToUint(customFieldId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			customFieldsRepository := repositories.NewCustomFieldRepository(nil)
			customField, err := customFieldsRepository.GetCustomFieldById(customFieldIdUint)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(customField)
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

func DeleteCustomField(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting custom field",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			customFieldId := chi.URLParam(r, "id")
			customFieldIdUint, err := utils.StringToUint(customFieldId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			customFieldsRepository := repositories.NewCustomFieldRepository(nil)
			err = customFieldsRepository.DeleteCustomField(customFieldIdUint)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
