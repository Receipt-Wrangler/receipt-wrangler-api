package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func tearDownUserTest() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToDeleteUser(t *testing.T) {
	defer tearDownUserTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "3")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.
		WithValue(
			r.Context(),
			jwtmiddleware.ContextKey{},
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}},
		)
	r = r.WithContext(newContext)

	DeleteUser(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToResetPassword(t *testing.T) {
	defer tearDownUserTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	newContext := context.
		WithValue(
			r.Context(),
			jwtmiddleware.ContextKey{},
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}},
		)
	r = r.WithContext(newContext)

	ResetPassword(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToConvertUser(t *testing.T) {
	defer tearDownUserTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	newContext := context.
		WithValue(
			r.Context(),
			jwtmiddleware.ContextKey{},
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}},
		)
	r = r.WithContext(newContext)

	ConvertDummyUserToNormalUser(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToCreateUser(t *testing.T) {
	defer tearDownUserTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	newContext := context.
		WithValue(
			r.Context(),
			jwtmiddleware.ContextKey{},
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}},
		)
	r = r.WithContext(newContext)

	CreateUser(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToUpdateUser(t *testing.T) {
	defer tearDownUserTest()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api", reader)
	var expectedStatusCode = http.StatusForbidden

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	newContext := context.
		WithValue(
			r.Context(),
			jwtmiddleware.ContextKey{},
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}},
		)
	r = r.WithContext(newContext)

	UpdateUser(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}
