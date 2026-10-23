package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
)

// putGroupSettings sends a raw JSON body, since a typed command cannot express
// an absent key.
func putGroupSettings(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api", strings.NewReader(body))

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("groupId", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiContext))
	r = r.WithContext(context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}}))

	UpdateGroupSettings(w, r)
	return w
}

func setUpGroupSettingsEmailTest(t *testing.T) {
	t.Helper()
	db := repositories.GetDB()
	db.Create(&models.Group{})
	db.Create(&models.GroupSettings{GroupId: 1})
	db.Create(&models.User{})
	grantAllAppPerms(t, 1)
}

func TestUpdateGroupSettingsRejectsAnInvalidSubjectRegex(t *testing.T) {
	defer tearDownGroupTests()
	setUpGroupSettingsEmailTest(t)

	w := putGroupSettings(t, `{"subjectLineRegexes": [{"regex": "^Receipt"}, {"regex": "^(?=Python-only)"}], "emailWhiteList": []}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "subjectLineRegexes.1.regex") {
		t.Errorf("body = %s, want the subjectLineRegexes.1.regex key", w.Body.String())
	}

	var stored []models.SubjectLineRegex
	repositories.GetDB().Find(&stored)
	if len(stored) != 0 {
		t.Errorf("a rejected save stored %d regexes", len(stored))
	}
}

func TestUpdateGroupSettingsEmailRequireSignedSenderRoundTrip(t *testing.T) {
	defer tearDownGroupTests()
	setUpGroupSettingsEmailTest(t)

	steps := []struct {
		name string
		body string
		want bool
	}{
		{"omitted key leaves the default", `{"emailWhiteList": [], "subjectLineRegexes": []}`, true},
		{"explicit false", `{"emailRequireSignedSender": false, "emailWhiteList": [], "subjectLineRegexes": []}`, false},
		{"omitted key leaves the stored false", `{"emailWhiteList": [], "subjectLineRegexes": []}`, false},
		{"explicit true", `{"emailRequireSignedSender": true, "emailWhiteList": [], "subjectLineRegexes": []}`, true},
	}

	for _, step := range steps {
		w := putGroupSettings(t, step.body)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, body = %s", step.name, w.Code, w.Body.String())
		}

		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response["emailRequireSignedSender"] != step.want {
			t.Errorf("%s: response emailRequireSignedSender = %v, want %v", step.name, response["emailRequireSignedSender"], step.want)
		}

		var stored models.GroupSettings
		repositories.GetDB().Where("group_id = ?", 1).First(&stored)
		if stored.EmailRequireSignedSender != step.want {
			t.Errorf("%s: stored = %v, want %v", step.name, stored.EmailRequireSignedSender, step.want)
		}
	}
}
