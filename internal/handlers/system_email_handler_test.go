package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func tearDownSystemEmailTest() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToCreateSystemEmail(t *testing.T) {
	defer tearDownSystemEmailTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	AddSystemEmail(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldCreateASystemEmail(t *testing.T) {
	defer tearDownSystemEmailTest()
	os.Setenv("ENCRYPTION_KEY", "superSecretKey")

	command := commands.UpsertSystemEmailCommand{
		Host:     "imap.gmail.com",
		Port:     "993",
		Username: "test@gmail.com",
		Password: "superSecretPassword",
	}
	bytes, _ := json.Marshal(command)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	AddSystemEmail(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldNotCreateASystemEmailDueToMissingEncryptionKey(t *testing.T) {
	defer tearDownSystemEmailTest()

	os.Setenv("ENCRYPTION_KEY", "")

	command := commands.UpsertSystemEmailCommand{
		Host:     "imap.gmail.com",
		Port:     "993",
		Username: "test@gmail.com",
		Password: "superSecretPassword",
	}
	bytes, _ := json.Marshal(command)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	AddSystemEmail(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldNotAllowUserToCreateInvalidSystemEmail(t *testing.T) {
	defer tearDownSystemEmailTest()

	tests := map[string]struct {
		input  commands.UpsertSystemEmailCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty test": {
			input:  commands.UpsertSystemEmailCommand{},
			expect: http.StatusBadRequest,
		},
		"missing host": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "",
				Port:     "993",
				Username: "test@gmail.com",
				Password: "superSecretPassword",
			},
			expect: http.StatusBadRequest,
		},
		"missing port": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "",
				Username: "username",
				Password: "password",
			},
			expect: http.StatusBadRequest,
		},
		"missing username": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "993",
				Username: "",
				Password: "password",
			},
			expect: http.StatusBadRequest,
		},
		"missing password": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "993",
				Username: "username",
				Password: "",
			},
			expect: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		AddSystemEmail(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}

func TestShouldNotAllowUserToGetSystemEmails(t *testing.T) {
	defer tearDownSystemEmailTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldReturn500WithMalformedPagedRequestCommand(t *testing.T) {
	defer tearDownSystemEmailTest()

	tests := map[string]struct {
		input  commands.PagedRequestCommand
		expect int
	}{
		"badOrderBy": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      50,
				OrderBy:       "badOrderBy",
				SortDirection: "asc",
			},
			expect: http.StatusInternalServerError,
		},
		"badSortDirection": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      50,
				OrderBy:       "host",
				SortDirection: "badSortDirection",
			},
			expect: http.StatusBadRequest,
		},
		"badPage": {
			input: commands.PagedRequestCommand{
				Page:          -1,
				PageSize:      50,
				OrderBy:       "host",
				SortDirection: "asc",
			},
			expect: http.StatusBadRequest,
		},
		"badPageSize": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      -2,
				OrderBy:       "host",
				SortDirection: "asc",
			},
			expect: http.StatusBadRequest,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		GetAllSystemEmails(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}
	}
}

func TestShouldAllowUserToGetSystemEmails(t *testing.T) {
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()

	pagedRequestCommand := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      50,
		OrderBy:       "host",
		SortDirection: "asc",
	}
	bytes, _ := json.Marshal(pagedRequestCommand)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	systemEmail := models.SystemEmail{
		BaseModel: models.BaseModel{},
		Host:      "imap.gmail.com",
		Port:      "993",
		Username:  "test@gmail.com",
		Password:  "superSecretPassword",
	}
	db.Create(&systemEmail)

	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	pagedData := structs.PagedData{}
	response, _ := io.ReadAll(w.Body)
	json.Unmarshal(response, &pagedData)

	if pagedData.TotalCount != 1 {
		utils.PrintTestError(t, pagedData.TotalCount, 1)
	}

	if pagedData.Data[0].(map[string]interface{})["host"] != "imap.gmail.com" {
		utils.PrintTestError(t, pagedData.Data[0].(map[string]interface{})["host"], "imap.gmail.com")
	}
}

func TestShouldNotGetSystemByIdAsUser(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	AddSystemEmail(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotGetSystemEmailByIdDueToBadId(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/systemEmail/{id}", reader)

	var expectedStatusCode = http.StatusInternalServerError

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetSystemEmailById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldGetSystemEmailById(t *testing.T) {
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/systemEmail/{id}", reader)

	var expectedStatusCode = http.StatusOK

	db.Create(&models.SystemEmail{
		Host:     "imap.gmail.com",
		Port:     "993",
		Username: "test",
		Password: "password",
	})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetSystemEmailById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotUpdateSystemEmailByIdAsAUser(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdateSystemEmail(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToUpdateInvalidSystemEmail(t *testing.T) {
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	tests := map[string]struct {
		input  commands.UpsertSystemEmailCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty test": {
			input:  commands.UpsertSystemEmailCommand{},
			expect: http.StatusBadRequest,
		},
		"missing host": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "",
				Port:     "993",
				Username: "test@gmail.com",
				Password: "superSecretPassword",
			},
			expect: http.StatusBadRequest,
		},
		"missing port": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "",
				Username: "username",
				Password: "password",
			},
			expect: http.StatusBadRequest,
		},
		"missing username": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "993",
				Username: "",
				Password: "password",
			},
			expect: http.StatusBadRequest,
		},
		"missing password": {
			input: commands.UpsertSystemEmailCommand{
				Host:     "imap.gmail.com",
				Port:     "993",
				Username: "username",
				Password: "",
			},
			expect: http.StatusOK,
		},
	}

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		chiContext := chi.NewRouteContext()
		chiContext.URLParams.Add("id", "1")
		routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
		r = r.WithContext(routeContext)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		UpdateSystemEmail(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}

func TestShouldNotDeleteEmailByIdAsAUser(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	DeleteSystemEmail(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotDeleteEmailWithBadId(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusInternalServerError

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeleteSystemEmail(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldDeleteEmail(t *testing.T) {
	defer tearDownSystemEmailTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusOK

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeleteSystemEmail(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserCheckConnectivity(t *testing.T) {
	defer tearDownSystemEmailTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CheckConnectivity(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowCheckInvalidConnectivityCommand(t *testing.T) {
	defer tearDownSystemEmailTest()

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	tests := map[string]struct {
		input  commands.CheckEmailConnectivityCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty test": {
			input:  commands.CheckEmailConnectivityCommand{},
			expect: http.StatusBadRequest,
		},
		"missing host": {
			input: commands.CheckEmailConnectivityCommand{
				UpsertSystemEmailCommand: commands.UpsertSystemEmailCommand{
					Host:     "",
					Port:     "993",
					Username: "test@gmail.com",
					Password: "superSecretPassword",
				},
			},
			expect: http.StatusBadRequest,
		},
		"missing port": {
			input: commands.CheckEmailConnectivityCommand{
				UpsertSystemEmailCommand: commands.UpsertSystemEmailCommand{
					Host:     "host",
					Port:     "",
					Username: "test@gmail.com",
					Password: "superSecretPassword",
				},
			},
			expect: http.StatusBadRequest,
		},
		"missing username": {
			input: commands.CheckEmailConnectivityCommand{
				UpsertSystemEmailCommand: commands.UpsertSystemEmailCommand{
					Host:     "host",
					Port:     "993",
					Username: "",
					Password: "superSecretPassword",
				},
			},
			expect: http.StatusBadRequest,
		},
		"missing password": {
			input: commands.CheckEmailConnectivityCommand{
				UpsertSystemEmailCommand: commands.UpsertSystemEmailCommand{
					Host:     "host",
					Port:     "993",
					Username: "test@gmail.com",
					Password: "",
				},
			},
			expect: http.StatusBadRequest,
		},
		"complete credentials": {
			input: commands.CheckEmailConnectivityCommand{
				UpsertSystemEmailCommand: commands.UpsertSystemEmailCommand{
					Host:     "host",
					Port:     "993",
					Username: "test@gmail.com",
					Password: "password",
				},
			},
			expect: http.StatusOK,
		},
		"just id": {
			input: commands.CheckEmailConnectivityCommand{
				ID: 1,
			},
			expect: http.StatusInternalServerError,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		CheckConnectivity(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s test expected: %d", name, test.expect))
		}
	}
}
