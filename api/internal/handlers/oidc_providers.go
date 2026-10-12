package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/oidc"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func GetPagedOidcProviders(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error getting OIDC providers",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppOidcProvidersRead},
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

			providers, count, err := repositories.NewOidcProviderRepository(nil).GetPagedOidcProviders(pagedRequestCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			views := services.BuildOidcProviderViews(providers)
			anyData := make([]any, len(views))
			for i := range views {
				anyData[i] = views[i]
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

func GetOidcProviderById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error getting OIDC provider",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppOidcProvidersRead},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id, err := utils.StringToUint(chi.URLParam(r, "oidcProviderId"))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			provider, err := repositories.NewOidcProviderRepository(nil).GetOidcProviderById(id)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return http.StatusNotFound, err
				}

				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(services.BuildOidcProviderView(provider))
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

func CreateOidcProvider(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error creating OIDC provider",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppOidcProvidersCreate},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			command := commands.UpsertOidcProviderCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(true)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			view, err := services.NewOidcProviderService(nil).CreateOidcProvider(command, &token.UserId)
			if err != nil {
				if handled := writeOidcProviderServiceError(w, err); handled {
					return 0, nil
				}

				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(view)
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

func UpdateOidcProvider(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error updating OIDC provider",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppOidcProvidersUpdate},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id, err := utils.StringToUint(chi.URLParam(r, "oidcProviderId"))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			command := commands.UpsertOidcProviderCommand{}
			err = command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(false)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			repository := repositories.NewOidcProviderRepository(nil)
			existing, err := repository.GetOidcProviderById(id)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return http.StatusNotFound, err
				}

				return http.StatusInternalServerError, err
			}

			vErrs = command.ValidateUpdate(existing)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			view, err := services.NewOidcProviderService(nil).UpdateOidcProvider(id, command)
			if err != nil {
				if handled := writeOidcProviderServiceError(w, err); handled {
					return 0, nil
				}

				return http.StatusInternalServerError, err
			}

			// Drop the cached discovery so an issuer or client-id change takes effect
			// immediately on this process. The UpdatedAt fingerprint already covers
			// correctness (including on other replicas); this just frees it promptly.
			oidc.InvalidateProvider(id)

			bytes, err := json.Marshal(view)
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

func DeleteOidcProvider(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error deleting OIDC provider",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppOidcProvidersDelete},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id, err := utils.StringToUint(chi.URLParam(r, "oidcProviderId"))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = repositories.NewOidcProviderRepository(nil).DeleteOidcProvider(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			oidc.InvalidateProvider(id)

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

// writeOidcProviderServiceError maps the service's typed errors onto a 400 with
// the same field-keyed shape a command validator produces, so the client renders
// them the same way. Reports whether it handled the error.
func writeOidcProviderServiceError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, services.ErrOidcProviderNameTaken):
		structs.WriteValidatorErrorResponse(w, structs.ValidatorError{
			Errors: map[string]string{"name": err.Error()},
		}, http.StatusBadRequest)

		return true
	case errors.Is(err, services.ErrServerPublicUrlRequired):
		structs.WriteValidatorErrorResponse(w, structs.ValidatorError{
			Errors: map[string]string{"enabled": err.Error()},
		}, http.StatusBadRequest)

		return true
	}

	return false
}
