package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestShouldReturn500WithMalformedPagedRequestCommand1(t *testing.T) {
	defer tearDownSystemEmailTest()

	pagedRequestCommand := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      50,
		OrderBy:       "badOrderBy",
		SortDirection: "asc",
	}
	bytes, _ := json.Marshal(pagedRequestCommand)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)
	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldReturn500WithMalformedPagedRequestCommand2(t *testing.T) {
	defer tearDownSystemEmailTest()

	pagedRequestCommand := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      50,
		OrderBy:       "host",
		SortDirection: "badSortDirection",
	}
	bytes, _ := json.Marshal(pagedRequestCommand)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)
	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusBadRequest)
	}
}

func TestShouldReturn500WithMalformedPagedRequestCommand3(t *testing.T) {
	defer tearDownSystemEmailTest()

	pagedRequestCommand := commands.PagedRequestCommand{
		Page:          -1,
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
	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusBadRequest)
	}
}

func TestShouldReturn500WithMalformedPagedRequestCommand4(t *testing.T) {
	defer tearDownSystemEmailTest()

	pagedRequestCommand := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      -1,
		OrderBy:       "host",
		SortDirection: "asc",
	}
	bytes, _ := json.Marshal(pagedRequestCommand)

	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)
	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusBadRequest)
	}
}

func TestShouldAllowUserToGetEmptySystemEmails(t *testing.T) {
	defer tearDownSystemEmailTest()

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

	repositories.CreateTestUser()

	GetAllSystemEmails(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	pagedData := structs.PagedData{}
	response, _ := io.ReadAll(w.Body)
	json.Unmarshal(response, &pagedData)

	if pagedData.TotalCount != 0 {
		utils.PrintTestError(t, pagedData.TotalCount, 0)
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
