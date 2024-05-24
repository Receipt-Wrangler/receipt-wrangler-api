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

	UpdateGroupSettings(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusInternalServerError)
	}
}

func TestShouldNotUserToUpdateGroupSettingsById(t *testing.T) {
	defer tearDownPromptTests()
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
	defer tearDownSystemEmailTest()
	db := repositories.GetDB()

	db.Create(&models.SystemEmail{})
	db.Create(&models.Group{})
	db.Create(&models.GroupSettings{
		GroupId: 1,
	})
	db.Create(&models.User{})
	id := uint(1)

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
		"valid command": {
			expect: http.StatusOK,
			input: commands.UpdateGroupSettingsCommand{
				EmailIntegrationEnabled:     true,
				SystemEmailId:               &id,
				EmailDefaultReceiptPaidById: &id,
				EmailDefaultReceiptStatus:   models.DRAFT,
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
