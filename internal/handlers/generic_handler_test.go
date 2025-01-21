package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func tearDownGenericHandlerTest() {
	repositories.TruncateTestDb()
}

func TestShouldSetContentTypeHeader(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	contentType := w.Header().Get("Content-Type")

	if contentType != constants.ApplicationJson {
		utils.PrintTestError(t, contentType, constants.ApplicationJson)
	}
}

func TestShouldRejectAccessBasedOnGroupId(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupId:      "2",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldRejectAccessBasedOnGroupIdIfGroupDoesNotExist(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupId:      "500",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldRejectAccessBasedOnGroupIdIfGroupIdIsMalformed(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupId:      "bad parse",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldRejectReceiptAccessBasedOnGroup(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()
	db := repositories.GetDB()
	receipt := models.Receipt{
		Name:         "Test receipt",
		GroupId:      1,
		PaidByUserID: 1,
	}
	db.Create(&receipt)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.EDITOR,
		ReceiptId:    "1",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldAcceptReceiptAccessBasedOnGroup(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()
	db := repositories.GetDB()
	receipt := models.Receipt{
		Name:         "Test receipt",
		GroupId:      1,
		PaidByUserID: 1,
	}
	db.Create(&receipt)

	db.Table("group_members").Where("user_id = ? & group_id = ?", 1, 1).Update("group_role", models.OWNER)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		ReceiptId:    "1",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldAcceptAccessBasedOnGroupId(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupId:      "1",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldAcceptReceiptsAccessBasedOnGroup(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()
	db := repositories.GetDB()
	receipt := models.Receipt{
		Name:         "Test receipt",
		GroupId:      1,
		PaidByUserID: 1,
	}
	receipt2 := models.Receipt{
		Name:         "Test receipt 2",
		GroupId:      1,
		PaidByUserID: 1,
	}
	db.Create(&receipt)
	db.Create(&receipt2)

	db.Table("group_members").Where("user_id = ? & group_id = ?", 1, 1).Update("group_role", models.OWNER)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.OWNER,
		ReceiptIds:   []string{"1", "2"},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldRejectAccessBasedOnEmptyGroupId(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupId:      "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldRejectAccessBasedOnEmptyGroupIds(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.VIEWER,
		GroupIds:     []string{},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

func TestShouldRejectReceiptAccessBasedOnWrongGroupRole(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()
	db := repositories.GetDB()
	receipt := models.Receipt{
		Name:         "Test receipt",
		GroupId:      1,
		PaidByUserID: 1,
	}
	db.Create(&receipt)

	db.Table("group_members").Where("user_id = ? & group_id = ?", 1, 1).Update("group_role", models.VIEWER)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.OWNER,
		ReceiptId:    "1",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldRejectReceiptAccessBasedOnWrongGroupRoleForMultipleReceipts(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	repositories.CreateTestGroupWithUsers()
	db := repositories.GetDB()
	receipt := models.Receipt{
		Name:         "Test receipt",
		GroupId:      1,
		PaidByUserID: 1,
	}
	receipt2 := models.Receipt{
		Name:         "Test receipt",
		GroupId:      2,
		PaidByUserID: 1,
	}
	db.Create(&receipt)
	db.Create(&receipt2)

	db.Table("group_members").Where("user_id = ? & group_id = ?", 1, 1).Update("group_role", models.OWNER)

	handler := structs.Handler{
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupRole:    models.OWNER,
		ReceiptIds:   []string{"1", "2"},
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldAcceptIfGroupIsAll(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2}})
	r = r.WithContext(newContext)

	handler := structs.Handler{
		Writer:  w,
		Request: r,
		GroupId: "all",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}
}

// TODO: Fix
// func TestShouldAcceptBasedOnUserRole(t *testing.T) {
// 	reader := strings.NewReader("")
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/api", reader)
// 	db := repositories.GetDB()

// 	db.Model(models.User{}).Where("id = ?", 2).Update("user_role", models.USER)

// 	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2}})
// 	r = r.WithContext(newContext)

// 	handler := structs.Handler{
// 		Writer:   w,
// 		Request:  r,
// 		UserRole: models.USER,
// 		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
// 			return 0, nil
// 		},
// 	}

// 	HandleRequest(handler)

// 	if w.Result().StatusCode != http.StatusOK {
// 		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
// 	}

// 	tearDownGenericHandlerTest()
// }

func TestShouldRejectBasedOnUserRole(t *testing.T) {
	defer tearDownGenericHandlerTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api", reader)
	db := repositories.GetDB()

	db.Model(models.User{}).Where("id = ?", 2).Update("user_role", models.USER)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 2}})
	r = r.WithContext(newContext)

	handler := structs.Handler{
		Writer:   w,
		Request:  r,
		UserRole: models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		},
	}

	HandleRequest(handler)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}
