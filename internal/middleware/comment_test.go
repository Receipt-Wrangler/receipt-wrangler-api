package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

func setupCommentMiddlewareTest() (http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	// Create User
	user := models.User{
		Username:    "test boy",
		Password:    "Password",
		DisplayName: "test",
	}
	db := db.GetDB()
	db.Create(&user)

	// Set up request
	reader := strings.NewReader("")
	r := httptest.NewRequest(http.MethodGet, "/api/1", reader)
	w := httptest.NewRecorder()

	var vClaims validator.ValidatedClaims
	vClaims.CustomClaims = &utils.Claims{UserId: 1}

	ctx := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &vClaims)
	r = r.WithContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), r, w
}

func teardownCommentMiddlewareTest() {
	db := db.GetDB()
	utils.TruncateTable(db, "comments")
	utils.TruncateTable(db, "receipts")
	utils.TruncateTable(db, "groups")
	utils.TruncateTable(db, "users")
}

func TestCanDeleteComment(t *testing.T) {
	// Define user id
	var userId uint
	userId = 1

	// Set up the request context
	fakeHandler, r, w := setupCommentMiddlewareTest()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("commentId", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Create test data
	db := db.GetDB()

	group := models.Group{}
	db.Create(&group)

	receipt := models.Receipt{
		Name:         "Name",
		Amount:       decimal.NewFromInt(100),
		Date:         time.Time{},
		GroupId:      group.ID,
		PaidByUserID: userId,
	}
	db.Create(&receipt)

	comment := models.Comment{
		Comment:   "Hello world",
		ReceiptId: receipt.ID,
		UserId:    &userId,
	}
	db.Create(&comment)

	// Serve
	handler := CanDeleteComment(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownCommentMiddlewareTest()
	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
	}
}

func TestCantDeleteComment(t *testing.T) {
	// Define user id
	var userId uint
	userId = 2

	// Set up the request context
	fakeHandler, r, w := setupCommentMiddlewareTest()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("commentId", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Create test data
	db := db.GetDB()

	// Create sescond user
	user := models.User{
		Username:    "test boy 2",
		Password:    "Password",
		DisplayName: "test",
	}
	db.Create(&user)

	group := models.Group{}
	db.Create(&group)

	receipt := models.Receipt{
		Name:         "Name",
		Amount:       decimal.NewFromInt(100),
		Date:         time.Time{},
		GroupId:      group.ID,
		PaidByUserID: userId,
	}
	db.Create(&receipt)

	comment := models.Comment{
		Comment:   "Hello world",
		ReceiptId: receipt.ID,
		UserId:    &userId,
	}
	db.Create(&comment)

	// Serve
	handler := CanDeleteComment(fakeHandler)
	handler.ServeHTTP(w, r)

	teardownCommentMiddlewareTest()
	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}
