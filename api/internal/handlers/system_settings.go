package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"receipt-wrangler/api/internal/wranglerasynq"
	"strings"
)

func GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error getting system settings",
		Writer:         w,
		Request:        r,
		AppPermissions: []string{permissions.AppSystemSettingsRead},
		ResponseType:   constants.ApplicationJson,
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
		ErrorMessage:   "Error updating system settings",
		Writer:         w,
		Request:        r,
		AppPermissions: []string{permissions.AppSystemSettingsUpdate},
		ResponseType:   constants.ApplicationJson,
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

			blocked, err := serverPublicUrlStillNeeded(command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if blocked {
				structs.WriteValidatorErrorResponse(w, structs.ValidatorError{
					Errors: map[string]string{
						"serverPublicUrl": services.ErrServerPublicUrlRequired.Error(),
					},
				}, http.StatusBadRequest)

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

// serverPublicUrlStillNeeded reports whether this save would clear the server
// public URL out from under an enabled OIDC provider.
//
// OidcProviderService.validateCrossSettingRules refuses to ENABLE a provider
// without the URL, but it only runs on a provider save -- so the same
// configuration was reachable from the other direction, by enabling a provider
// and then blanking the setting. The failure is silent and lands somewhere
// nobody looks: GetServerPublicUrl falls back to http://localhost:8081, every
// provider starts sending a redirect_uri that was never registered, and the
// rejection happens at the identity provider, where this API never sees it.
//
// The trigger is a blank value, deliberately the same test
// IsServerPublicUrlConfigured applies, so the two halves of the rule cannot
// disagree about what "configured" means. CHANGING the URL stays allowed -- a
// deployment legitimately moves, and the admin re-registers the redirect URI --
// and a nil pointer means the key was omitted, which leaves the stored value
// alone and can never break anything.
func serverPublicUrlStillNeeded(command commands.UpsertSystemSettingsCommand) (bool, error) {
	if command.ServerPublicUrl == nil || len(strings.TrimSpace(*command.ServerPublicUrl)) > 0 {
		return false, nil
	}

	enabledProviders, err := repositories.NewOidcProviderRepository(nil).GetAllEnabledOidcProviders()
	if err != nil {
		return false, err
	}

	return len(enabledProviders) > 0, nil
}

func RestartTaskServer(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error restarting task server, please restart the entire application",
		Writer:         w,
		Request:        r,
		AppPermissions: []string{permissions.AppSystemSettingsRestartTaskServer},
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
