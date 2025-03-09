package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
)

func setupCategoriesTest() {
	repositories.CreateTestCategories()
}

func tearDownCategoriesTest() {
	repositories.TruncateTestDb()
}

func TestShouldGetAllCategories(t *testing.T) {
	defer tearDownCategoriesTest()
	categories := make([]models.Category, 0)
	setupCategoriesTest()

	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetAllCategories(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &categories)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}

	for index, category := range categories {
		id := index + 1
		if uint(id) != category.ID {
			utils.PrintTestError(t, category.ID, id)
		}
	}

	tearDownCategoriesTest()
}

func TestShouldCreateCategory(t *testing.T) {
	defer tearDownCategoriesTest()
	category := models.Category{}
	setupCategoriesTest()

	reader := strings.NewReader(`{"name": "Test category", "description": "Test description"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CreateCategory(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &category)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}

	if category.Name != "Test category" {
		utils.PrintTestError(t, category.Name, "Test category")
	}

	if category.Description != "Test description" {
		utils.PrintTestError(t, category.Description, "Test description")
	}

	if category.ID != 4 {
		utils.PrintTestError(t, category.ID, 4)
	}
}

func TestShouldUpdateCategoryIfAdmin(t *testing.T) {
	defer tearDownCategoriesTest()
	category := models.Category{}
	setupCategoriesTest()

	reader := strings.NewReader(`{"name": "Updated Category name", "description": "Updated Test description"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryId", "1")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	UpdateCategory(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &category)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if category.Name != "Updated Category name" {
		utils.PrintTestError(t, category.Name, "Updated Category name")
	}

	if category.Description != "Updated Test description" {
		utils.PrintTestError(t, category.Description, "Updated Test description")
	}
}

func TestShouldNotUpdateCategoryDueToRole(t *testing.T) {
	defer tearDownCategoriesTest()
	category := models.Category{}
	setupCategoriesTest()

	reader := strings.NewReader(`{"name": "Updated Category name", "description": "Updated Test description"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryId", "1")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdateCategory(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &category)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestShouldDeleteCategoryIfAdmin(t *testing.T) {
	defer tearDownCategoriesTest()
	setupCategoriesTest()

	reader := strings.NewReader(``)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryId", "1")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeleteCategory(w, r)

	db := repositories.GetDB()
	err := db.Model(models.Category{}).Where("id = ?", 1).First(&models.Category{}).Error
	if err == nil {
		utils.PrintTestError(t, err, "Record should not exist")
	}

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestShouldNotDeleteCategoryDueToRole(t *testing.T) {
	defer tearDownCategoriesTest()
	setupCategoriesTest()

	reader := strings.NewReader(``)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryId", "1")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.USER}})
	r = r.WithContext(newContext)

	DeleteCategory(w, r)

	if w.Result().StatusCode != 403 {
		utils.PrintTestError(t, w.Result().StatusCode, 403)
	}
}

func TestShouldGetCategoryNameCountIfAdmin(t *testing.T) {
	defer tearDownCategoriesTest()
	expectedStatus := 200
	var resultCount uint
	setupCategoriesTest()

	reader := strings.NewReader(``)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryName", "test")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetCategoryNameCount(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &resultCount)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != expectedStatus {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatus)
	}

	if resultCount != 1 {
		utils.PrintTestError(t, resultCount, 1)
	}
}

func TestShouldGetCategoryNameCountIfAdmin2(t *testing.T) {
	defer tearDownCategoriesTest()
	expectedStatus := 200
	var resultCount uint
	setupCategoriesTest()

	reader := strings.NewReader(``)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryName", "totally a category name")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetCategoryNameCount(w, r)

	err := json.Unmarshal(w.Body.Bytes(), &resultCount)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != expectedStatus {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatus)
	}

	if resultCount != 0 {
		utils.PrintTestError(t, resultCount, 0)
	}
}

func TestShouldNotGetCategoryNameDueToRole(t *testing.T) {
	defer tearDownCategoriesTest()
	expectedStatus := 403
	setupCategoriesTest()

	reader := strings.NewReader(``)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("categoryName", "totally a category name")

	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, ctx)
	r = r.WithContext(routeContext)
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetCategoryNameCount(w, r)

	if w.Result().StatusCode != expectedStatus {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatus)
	}
}
