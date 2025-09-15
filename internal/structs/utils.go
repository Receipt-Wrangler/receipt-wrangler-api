package structs

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/utils"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func GetClaims(r *http.Request) *Claims {
	return r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims).CustomClaims.(*Claims)
}

func WriteValidatorErrorResponse(w http.ResponseWriter, err ValidatorError, responseCode int) {
	bytes, marshalErr := json.Marshal(err.Errors)
	if marshalErr != nil {
		utils.WriteErrorResponse(w, marshalErr, responseCode)
	}

	logging.LogStd(logging.LOG_LEVEL_ERROR, string(bytes))

	w.WriteHeader(responseCode)
	w.Write(bytes)
}
