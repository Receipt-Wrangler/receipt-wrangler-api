package httpUtils

import (
	"io/ioutil"
	"net/http"
)

func GetBodyData(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	return bodyData, err
}

func WriteErrorResponse(w http.ResponseWriter, err error, responseCode int) {
	w.WriteHeader(responseCode)
	w.Write([]byte(err.Error()))
}
