package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"receipt-wrangler/api/internal/structs"
)

var errKey = "errorMsg"

func GetBodyData(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	return bodyData, err
}

func WriteErrorResponse(w http.ResponseWriter, err error, responseCode int) {
	errMap := make(map[string]string)
	errMap[errKey] = err.Error()

	bytes, marshalErr := json.Marshal(errMap)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func WriteCustomErrorResponse(w http.ResponseWriter, msg string, responseCode int) {
	errMap := make(map[string]string)
	errMap[errKey] = msg

	bytes, marshalErr := json.Marshal(errMap)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, 500)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func WriteValidatorErrorResponse(w http.ResponseWriter, err structs.ValidatorError, responseCode int) {
	bytes, marshalErr := json.Marshal(err.Errors)
	if marshalErr != nil {
		WriteErrorResponse(w, marshalErr, responseCode)
	}

	w.WriteHeader(responseCode)
	w.Write(bytes)
}

func MarshalResponseData(data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func SetJSONResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
