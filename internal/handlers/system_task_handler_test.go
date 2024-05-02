package handlers

import (
	"context"
	"encoding/json"
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

func tearDownSystemTaskTest() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToGetSystemTasks(t *testing.T) {
	defer tearDownSystemTaskTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetSystemTasks(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotAllowAdminToGetSystemTasksWithInvalidCommand(t *testing.T) {
	defer tearDownSystemTaskTest()

	tests := map[string]struct {
		input  commands.GetSystemTaskCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty command": {
			input:  commands.GetSystemTaskCommand{},
			expect: http.StatusBadRequest,
		},
		"missing count": {
			input: commands.GetSystemTaskCommand{
				AssociatedEntityId:   1,
				AssociatedEntityType: models.SYSTEM_EMAIL,
			},
			expect: http.StatusOK,
		},
		"missing associated entityId": {
			input: commands.GetSystemTaskCommand{
				AssociatedEntityType: models.SYSTEM_EMAIL,
				Count:                10,
			},
			expect: http.StatusOK,
		},
		"missing associated entityType": {
			input: commands.GetSystemTaskCommand{
				AssociatedEntityId: 1,
				Count:              10,
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

		GetSystemTasks(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}
