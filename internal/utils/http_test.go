package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TODO: Fix
func TestGetBodyDataGetsData(t *testing.T) {
	var unmarshalResult any
	testString := "my test string wowzer"
	reader := strings.NewReader(testString)
	r := httptest.NewRequest(http.MethodGet, "/api", reader)
	w := httptest.NewRecorder()
	bytes, _ := GetBodyData(w, r)

	json.Unmarshal(bytes, &unmarshalResult)

	if testString != unmarshalResult {
		// repositories.PrintTestError(t, unmarshalResult, testString)
	}
}

func TestWriteErrorResponseWritesReponse(t *testing.T) {
	var errBytes = make([]byte, 100)
	var errMap map[string]string

	w := httptest.NewRecorder()
	err := fmt.Errorf("Test error")

	WriteErrorResponse(w, err, 500)

	if w.Result().StatusCode != 500 {
		PrintTestError(t, w.Result().StatusCode, 500)
	}

	w.Body.Read(errBytes)
	json.Unmarshal(errBytes[0:25], &errMap)

	if errMap[errKey] != "Test error" {
		PrintTestError(t, errMap[errKey], "Test error")
	}
}

func TestWriteCustomErrorResponseWritesResponse(t *testing.T) {
	var errBytes = make([]byte, 100)
	var errMap map[string]string

	customMsg := "Hello world"

	w := httptest.NewRecorder()

	WriteCustomErrorResponse(w, customMsg, 200)

	if w.Result().StatusCode != 200 {
		PrintTestError(t, w.Result().StatusCode, 200)
	}

	w.Body.Read(errBytes)
	json.Unmarshal(errBytes[0:26], &errMap)

	if errMap[errKey] != customMsg {
		PrintTestError(t, errMap[errKey], customMsg)
	}
}

// TODO: move test
// func TestWriteValidatorErrorResponseWritesResponse(t *testing.T) {
// 	var errBytes = make([]byte, 100)
// 	var bodyVErr structs.ValidatorError
// 	vErr := structs.ValidatorError{
// 		Errors: make(map[string]string),
// 	}
// 	nameErr := "error"
// 	amountErr := "amount cannot be empty"

// 	vErr.Errors["name"] = nameErr
// 	vErr.Errors["amount"] = amountErr

// 	w := httptest.NewRecorder()

// 	WriteValidatorErrorResponse(w, vErr, 400)

// 	if w.Result().StatusCode != 400 {
// 		PrintTestError(t, w.Result().StatusCode, 400)
// 	}

// 	w.Body.Read(errBytes)
// 	json.Unmarshal(errBytes[0:50], &bodyVErr)

// 	if reflect.DeepEqual(vErr, bodyVErr) {
// 		PrintTestError(t, vErr, bodyVErr)
// 	}
// }
