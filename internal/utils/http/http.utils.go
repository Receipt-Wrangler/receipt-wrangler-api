package httpUtils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	handlers "receipt-wrangler/api/internal/handlers"
)

func GetBodyData(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	return bodyData, err
}

func WriteErrorResponse(w http.ResponseWriter, err error, responseCode int) {
	w.WriteHeader(responseCode)
	w.Write([]byte(err.Error()))
}

func WriteCustomErrorResponse(w http.ResponseWriter, msg string, responseCode int) {
	w.WriteHeader(responseCode)
	w.Write([]byte(msg))
}

func WriteValidatorErrorResponse(w http.ResponseWriter, err handlers.ValidatorError, responseCode int) {

	bytes, marshalErr := json.Marshal(err.Errors)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}
