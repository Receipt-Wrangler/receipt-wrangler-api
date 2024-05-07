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
				PageSize:      -1,
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
