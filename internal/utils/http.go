package utils

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
	errMap := make(map[string]string)
	errMap["errorMsg"] = err.Error()

	bytes, marshalErr := json.Marshal(errMap)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func WriteCustomErrorResponse(w http.ResponseWriter, msg string, responseCode int) {
	errMap := make(map[string]string)
	errMap["errorMsg"] = msg

	bytes, marshalErr := json.Marshal(errMap)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func WriteValidatorErrorResponse(w http.ResponseWriter, err handlers.ValidatorError, responseCode int) {
	bytes, marshalErr := json.Marshal(err.Errors)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}
