package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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

// OidcExchange redeems the mobile handoff code. Unauthenticated: the code plus
// its PKCE proof is the authentication.
func OidcExchange(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error completing sign in",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			appData, err := oidc.Exchange(w, r)
			if err != nil {
				// Every rejection reason is deliberately indistinguishable, so a caller
				// cannot probe which codes exist.
				if errors.Is(err, oidc.ErrInvalidExchange) {
					return http.StatusBadRequest, err
				}

				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(appData)
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

// GetOidcConnections lists the calling user's linked providers.
func GetOidcConnections(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error getting connected accounts",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppAccountRead},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			identities, err := repositories.NewOidcIdentityRepository(nil).GetIdentitiesForUser(token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(services.BuildOidcConnectionViews(identities))
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

// DeleteOidcConnection unlinks one of the caller's providers.
func DeleteOidcConnection(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage:   "Error disconnecting account",
		Writer:         w,
		Request:        r,
		ResponseType:   constants.ApplicationJson,
		AppPermissions: []string{permissions.AppAccountUpdate},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			// Managing which identity providers can sign into an account is for the
			// account's own session, not for a key it issued. oidc.LinkStart refuses an
			// API key because adding a sign-in method escalates a scoped credential into
			// a permanent one; this endpoint carries the same permission for the stated
			// reason that removing one is no less privileged than adding it, so it
			// carries the same rule. Required-to-be-a-session, not
			// refused-if-a-key, so an unmarked caller is denied too.
			if token.AuthSource != structs.AuthSourceJwt {
				return http.StatusForbidden, errors.New("an OIDC connection may only be removed by an interactive session")
			}

			err := oidc.UnlinkIdentity(token.UserId, chi.URLParam(r, "name"))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return http.StatusNotFound, err
				}

				if errors.Is(err, oidc.ErrWouldLockOut) {
					structs.WriteValidatorErrorResponse(w, structs.ValidatorError{
						Errors: map[string]string{
							"providerName": "This account was created by this provider and has no password to sign in with. Ask an administrator to set one before disconnecting.",
						},
					}, http.StatusBadRequest)

					return 0, nil
				}

				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
