package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
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
)

func tearDownGroupTests() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowAdminToUpdateGroupSettingsIdWithBadId(t *testing.T) {
	defer tearDownGroupTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	UpdateGroupSettings(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldNotUserToUpdateGroupSettingsById(t *testing.T) {
	defer tearDownGroupTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api", reader)
	expectedStatusCode := http.StatusForbidden

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("groupId", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdateGroupReceiptSettings(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotUserToUpdateGroupReceiptSettingsById(t *testing.T) {
	defer tearDownGroupTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)
	expectedStatusCode := http.StatusForbidden

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdateGroupSettings(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldTestUpdateGroupSettingsWithVariousCommands(t *testing.T) {
	defer tearDownGroupTests()
	db := repositories.GetDB()

	db.Create(&models.SystemEmail{})
	db.Create(&models.Group{})
	db.Create(&models.GroupSettings{
		GroupId: 1,
	})
	db.Create(&models.User{})

	db.Create(&models.Prompt{})
	db.Create(&models.Prompt{})

	id := uint(1)
	badId := uint(0)

	tests := map[string]struct {
		input  commands.UpdateGroupSettingsCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusOK,
		},
		"empty command": {
			expect: http.StatusOK,
			input:  commands.UpdateGroupSettingsCommand{},
		},
		"bad enabled email integration, due to missing email id": {
			expect: http.StatusBadRequest,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled:     true,
				EmailDefaultReceiptPaidById: &id,
				EmailDefaultReceiptStatus:   models.DRAFT,
			},
		},
		"bad enabled email integration, due to missing default receipt paid by id": {
			expect: http.StatusBadRequest,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled:   true,
				SystemEmailId:             &id,
				EmailDefaultReceiptStatus: models.DRAFT,
			},
		},
		"bad enabled email integration, due to missing default receipt status": {
			expect: http.StatusBadRequest,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled:     true,
				SystemEmailId:               &id,
				EmailDefaultReceiptPaidById: &id,
			},
		},
		"bad enabled email integration, due to only including EmailIntegrationEnabled": {
			expect: http.StatusBadRequest,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled: true,
			},
		},
		"fallback prompt id without main id": {
			expect: http.StatusOK,
			input: commands.UpdateGroupSettingsCommand{
				FallbackPromptId: &id,
			},
		},
		"invalid prompt id": {
			expect: http.StatusBadRequest,
			input: commands.UpdateGroupSettingsCommand{
				PromptId: &badId,
			},
		},
		"valid command": {
			expect: http.StatusOK,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled:     true,
				SystemEmailId:               &id,
				EmailDefaultReceiptPaidById: &id,
				EmailDefaultReceiptStatus:   models.DRAFT,
				PromptId:                    &id,
				FallbackPromptId:            &id,
			},
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		chiContext := chi.NewRouteContext()
		chiContext.URLParams.Add("groupId", "1")
		routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
		r = r.WithContext(routeContext)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		UpdateGroupSettings(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected %d", name, test.expect))
		}
	}
}

func TestShouldTestUpdateGroupSettingsWithVariousCommandsAsAdmin(t *testing.T) {
	defer tearDownGroupTests()

	tests := map[string]struct {
		input    commands.PagedGroupRequestCommand
		expect   int
		userRole models.UserRole
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty command": {
			input:  commands.PagedGroupRequestCommand{},
			expect: http.StatusBadRequest,
		},
		"bad command due to bad orderBy": {
			input: commands.PagedGroupRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "badOrderBy",
					SortDirection: commands.ASCENDING,
				},
				GroupFilter: commands.GroupFilter{
					AssociatedGroup: commands.ASSOCIATED_GROUP_ALL,
				},
			},
			userRole: models.ADMIN,
			expect:   http.StatusInternalServerError,
		},
		"valid command with all groups as admin": {
			input: commands.PagedGroupRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				GroupFilter: commands.GroupFilter{
					AssociatedGroup: commands.ASSOCIATED_GROUP_ALL,
				},
			},
			expect:   http.StatusOK,
			userRole: models.ADMIN,
		},
		"bad command with all groups as user": {
			input: commands.PagedGroupRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				GroupFilter: commands.GroupFilter{
					AssociatedGroup: commands.ASSOCIATED_GROUP_ALL,
				},
			},
			expect:   http.StatusBadRequest,
			userRole: models.USER,
		},
		"valid command with my groups as admin": {
			input: commands.PagedGroupRequestCommand{
				PagedRequestCommand: commands.PagedRequestCommand{
					Page:          1,
					PageSize:      10,
					OrderBy:       "name",
					SortDirection: commands.ASCENDING,
				},
				GroupFilter: commands.GroupFilter{
					AssociatedGroup: commands.ASSOCIATED_GROUP_MINE,
				},
			},
			expect:   http.StatusOK,
			userRole: models.ADMIN,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: test.userRole}})
		r = r.WithContext(newContext)

		GetPagedGroups(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected %d", name, test.expect))
		}
	}
}
