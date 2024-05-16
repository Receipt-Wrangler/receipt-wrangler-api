package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
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

func tearDownPromptTests() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToGetPrompts(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetPagedPrompts(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotGetPagedPromptsWithBadRequest(t *testing.T) {
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
		"valid": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      25,
				OrderBy:       "name",
				SortDirection: "asc",
			},
			expect: http.StatusOK,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		GetPagedPrompts(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}
	}
}

func TestShouldNotAllowUserToGetPromptById(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetPromptById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowAdminToGetPromptByIdWithBadId(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetPromptById(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldAllowAdminToGetPromptById(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	db := repositories.GetDB()
	db.Create(&models.Prompt{})

	GetPromptById(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldNotAllowUserToUpdatePromptById(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdatePromptById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowAdminToUpdatePromptByIdWithBadId(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	UpdatePromptById(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldNotAllowUserToUpdateInvalidPrompt(t *testing.T) {
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()
	db.Create(&models.Prompt{})

	tests := map[string]struct {
		input  commands.UpsertPromptCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty test": {
			input:  commands.UpsertPromptCommand{},
			expect: http.StatusBadRequest,
		},
		"empty prompt": {
			input: commands.UpsertPromptCommand{
				Name: "test",
			},
			expect: http.StatusBadRequest,
		},
		"empty name": {
			input: commands.UpsertPromptCommand{
				Prompt: "test",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable in middle": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "hello this @bad is bad",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable at end": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "hello this @bad",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable at beginning": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "@bad variable",
			},
			expect: http.StatusBadRequest,
		},
		"bad template embedded variable": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "asdlfkasjdfldj@badvariablelaksjdfasldjk",
			},
			expect: http.StatusBadRequest,
		},
		"good prompt no variables": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "hello variable",
			},
			expect: http.StatusOK,
		},
		"good prompt with variables": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "please use, @categories, and @tags",
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

		UpdatePromptById(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}

func TestShouldNotAllowUserToCreatePrompt(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CreatePrompt(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowAdminToCreateInvalidPrompt(t *testing.T) {
	defer tearDownSystemEmailTest()

	tests := map[string]struct {
		input  commands.UpsertPromptCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty test": {
			input:  commands.UpsertPromptCommand{},
			expect: http.StatusBadRequest,
		},
		"empty prompt": {
			input: commands.UpsertPromptCommand{
				Name: "test",
			},
			expect: http.StatusBadRequest,
		},
		"empty name": {
			input: commands.UpsertPromptCommand{
				Prompt: "test",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable in middle": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "hello this @bad is bad",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable at end": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "hello this @bad",
			},
			expect: http.StatusBadRequest,
		},
		"bad template variable at beginning": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "@bad variable",
			},
			expect: http.StatusBadRequest,
		},
		"bad template embedded variable": {
			input: commands.UpsertPromptCommand{
				Name:   "test",
				Prompt: "asdlfkasjdfldj@badvariablelaksjdfasldjk",
			},
			expect: http.StatusBadRequest,
		},
		"good prompt no variables": {
			input: commands.UpsertPromptCommand{
				Name:   "test1",
				Prompt: "hello variable",
			},
			expect: http.StatusOK,
		},
		"good prompt with variables": {
			input: commands.UpsertPromptCommand{
				Name:   "test2",
				Prompt: "please use, @categories, and @tags",
			},
			expect: http.StatusOK,
		},
	}

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		CreatePrompt(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}

func TestShouldNotAllowUserToDeletePromptById(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	DeletePromptById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowAdminToDeletePromptWithBadId(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeletePromptById(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldDeletePromptById(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	db := repositories.GetDB()
	db.Create(&models.Prompt{})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeletePromptById(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldNotAllowUserToCreateDefaultPrompt(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CreateDefaultPrompt(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldCreateDefaultPrompt(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	expectedStatus := http.StatusOK

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	CreateDefaultPrompt(w, r)

	if w.Result().StatusCode != expectedStatus {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatus)
	}
}

func TestShouldNotAllowToCreateDuplicateDefaultPrompts(t *testing.T) {
	defer tearDownPromptTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	w2 := httptest.NewRecorder()

	expectedStatus := http.StatusInternalServerError

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	CreateDefaultPrompt(w, r)
	CreateDefaultPrompt(w2, r)

	if w2.Result().StatusCode != expectedStatus {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatus)
	}
}
