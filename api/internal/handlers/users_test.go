package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

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
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}},
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
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}},
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
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}},
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
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}},
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
			&validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}},
		)
	r = r.WithContext(newContext)

	UpdateUser(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func createTestUser(t *testing.T, username string, password string) models.User {
	userRepository := repositories.NewUserRepository(nil)
	user, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    username,
		DisplayName: username,
		Password:    password,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestDeleteAccountShouldFailWithWrongPassword(t *testing.T) {
	defer tearDownUserTest()

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	user := createTestUser(t, "testuser", "correctpassword")

	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/user/deleteAccount", reader)

	ctx := context.WithValue(r.Context(), "deleteAccountCommand", commands.DeleteAccountCommand{Password: "wrongpassword"})
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: user.ID, AuthSource: structs.AuthSourceJwt},
	})
	r = r.WithContext(ctx)

	grantAppPerms(t, user.ID, permissions.AppAccountDelete)

	DeleteAccount(w, r)

	if w.Result().StatusCode != http.StatusUnauthorized {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusUnauthorized)
	}
}

func TestDeleteAccountShouldSucceedWithCorrectPassword(t *testing.T) {
	defer tearDownUserTest()

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	// Create a second admin so this user is not the only admin
	adminUser := createTestUser(t, "adminuser", "adminpass")
	grantAppPerms(t, adminUser.ID, permissions.AppUsersRead)
	user := createTestUser(t, "testuser", "correctpassword")

	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/user/deleteAccount", reader)

	ctx := context.WithValue(r.Context(), "deleteAccountCommand", commands.DeleteAccountCommand{Password: "correctpassword"})
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: user.ID, AuthSource: structs.AuthSourceJwt},
	})
	r = r.WithContext(ctx)

	grantAppPerms(t, user.ID, permissions.AppAccountDelete)

	DeleteAccount(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
	}

	// Verify user was deleted
	var count int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 0 {
		utils.PrintTestError(t, count, 0)
	}
}

func TestDeleteAccountShouldPreventLastAdminDeletion(t *testing.T) {
	defer tearDownUserTest()

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	// This user will be the only admin. "Administrator" is defined by the
	// modern app.users.read permission (see services.DeleteAccount), so grant
	// it the admin permission alongside the self-service delete permission.
	user := createTestUser(t, "onlyadmin", "adminpassword")
	grantAppPerms(t, user.ID, permissions.AppUsersRead, permissions.AppAccountDelete)

	// Ensure no other admins exist
	roleRepository := repositories.NewRoleRepository(nil)
	adminCount, err := roleRepository.CountUsersWithAppPermission(permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("count admins: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("Expected exactly 1 admin, got %d", adminCount)
	}

	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/user/deleteAccount", reader)

	ctx := context.WithValue(r.Context(), "deleteAccountCommand", commands.DeleteAccountCommand{Password: "adminpassword"})
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: user.ID, AuthSource: structs.AuthSourceJwt},
	})
	r = r.WithContext(ctx)

	DeleteAccount(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusBadRequest)
	}

	// Verify user was NOT deleted
	var count int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 1 {
		utils.PrintTestError(t, count, 1)
	}
}

// deleteAccountAs drives POST /user/deleteAccount with claims built the way the
// named auth source would have built them.
func deleteAccountAs(t *testing.T, userId uint, password string, authSource string) *httptest.ResponseRecorder {
	t.Helper()

	r := httptest.NewRequest("POST", "/api/user/deleteAccount", strings.NewReader(""))

	ctx := context.WithValue(r.Context(), "deleteAccountCommand", commands.DeleteAccountCommand{Password: password})
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: userId, AuthSource: authSource},
	})

	w := httptest.NewRecorder()
	DeleteAccount(w, r.WithContext(ctx))

	return w
}

// TestDeleteAccountRequiresAnInteractiveSession pins self-delete to the account's
// own session. A key must not be able to destroy the account that issued it,
// whatever its scope -- the same rule connecting and disconnecting an identity
// provider carry.
func TestDeleteAccountRequiresAnInteractiveSession(t *testing.T) {
	defer tearDownUserTest()

	db := repositories.GetDB()
	db.Create(&models.SystemEmail{})

	// A second admin, so the last-admin guard cannot be what refuses the delete.
	adminUser := createTestUser(t, "adminuser", "adminpass")
	grantAppPerms(t, adminUser.ID, permissions.AppUsersRead)
	user := createTestUser(t, "testuser", "correctpassword")
	grantAppPerms(t, user.ID, permissions.AppAccountDelete)

	// Unmarked claims are denied too: the check requires an interactive session
	// rather than rejecting a key, so the zero value cannot slip through.
	for _, authSource := range []string{structs.AuthSourceApiKey, ""} {
		w := deleteAccountAs(t, user.ID, "correctpassword", authSource)
		if w.Result().StatusCode != http.StatusForbidden {
			t.Errorf("auth source %q: expected 403, got %d", authSource, w.Result().StatusCode)
		}
	}

	var count int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Fatalf("a refused self-delete removed the account anyway")
	}

	// The positive control: the same call from a real session succeeds, so the
	// 403s above are the credential rule and not some unrelated failure.
	w := deleteAccountAs(t, user.ID, "correctpassword", structs.AuthSourceJwt)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected an interactive session to delete the account, got %d", w.Result().StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GetAmountOwedForUser tests
// ---------------------------------------------------------------------------

func setupAmountOwedTest(t *testing.T) {
	repositories.CreateTestGroupWithUsers()
	// GetAmountOwedForUser is gated on group.receipts.read; authorize user 1 in
	// group 1 (the group the positive cases query).
	grantGroupPerms(t, 1, 1, permissions.GroupReceiptsRead)
}

func createReceiptWithItems(
	t *testing.T,
	name string,
	amount float64,
	paidByUserId uint,
	groupId uint,
	items []commands.UpsertItemCommand,
) models.Receipt {
	t.Helper()
	receiptRepository := repositories.NewReceiptRepository(nil)
	cmd := commands.UpsertReceiptCommand{
		Name:         name,
		Amount:       decimal.NewFromFloat(amount),
		Date:         time.Now(),
		PaidByUserID: paidByUserId,
		GroupId:      groupId,
		Status:       models.OPEN,
		Items:        items,
	}

	receipt, err := receiptRepository.CreateReceipt(cmd, paidByUserId, true)
	if err != nil {
		t.Fatalf("failed to create test receipt %q: %v", name, err)
	}
	return receipt
}

func chargedItem(name string, amount float64, chargedToUserId uint) commands.UpsertItemCommand {
	return commands.UpsertItemCommand{
		Name:            name,
		Amount:          decimal.NewFromFloat(amount),
		Status:          models.ITEM_OPEN,
		ChargedToUserId: uintPtr(chargedToUserId),
	}
}

func chargedItemWithStatus(name string, amount float64, chargedToUserId uint, status models.ItemStatus) commands.UpsertItemCommand {
	return commands.UpsertItemCommand{
		Name:            name,
		Amount:          decimal.NewFromFloat(amount),
		Status:          status,
		ChargedToUserId: uintPtr(chargedToUserId),
	}
}

func callGetAmountOwed(callerUserId uint, groupId string, receiptIds []string) (*httptest.ResponseRecorder, map[uint]decimal.Decimal) {
	form := url.Values{}
	for _, id := range receiptIds {
		form.Add("receiptIds", id)
	}

	target := "/api/user/getAmountOwedForUser"
	if groupId != "" {
		target += "?groupId=" + url.QueryEscape(groupId)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", target, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: callerUserId},
	})
	r = r.WithContext(ctx)

	GetAmountOwedForUser(w, r)

	if w.Result().StatusCode != http.StatusOK {
		return w, nil
	}

	// The handler marshals map[uint]decimal.Decimal — JSON object keys are strings,
	// so unmarshal into a string-keyed map and convert.
	stringKeyed := map[string]decimal.Decimal{}
	if err := json.Unmarshal(w.Body.Bytes(), &stringKeyed); err != nil {
		return w, nil
	}

	result := make(map[uint]decimal.Decimal, len(stringKeyed))
	for k, v := range stringKeyed {
		parsed, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		result[uint(parsed)] = v
	}
	return w, result
}

func assertOwed(t *testing.T, result map[uint]decimal.Decimal, otherUserId uint, expected float64) {
	t.Helper()
	exp := decimal.NewFromFloat(expected)
	got, ok := result[otherUserId]
	if !ok {
		t.Errorf("expected entry for user %d (=%s) but none found; result=%v", otherUserId, exp.String(), result)
		return
	}
	if !got.Equal(exp) {
		t.Errorf("expected resultMap[%d] == %s, got %s", otherUserId, exp.String(), got.String())
	}
}

// setupIsolatedAmountOwedTest seeds group 1 as an isolated group where user 1 is
// the (restricted) caller, user 3 is a visible supervisor, and user 2 is an
// invisible peer.
func setupIsolatedAmountOwedTest(t *testing.T) {
	repositories.CreateTestGroupWithUsers()
	grantGroupPerms(t, 1, 1, permissions.GroupReceiptsRead)
	isolateGroupWithSupervisor(t, 1, 3)
}

// --- Member isolation ---------------------------------------------------

func TestGetAmountOwedForUserIsolatedViewerExcludesInvisibleCounterparty(t *testing.T) {
	defer tearDownUserTest()
	setupIsolatedAmountOwedTest(t)

	// User 2 (invisible peer) paid; item charged to user 1 -> entry for user 2.
	createReceiptWithItems(t, "Peer paid", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("peer item", 10, 1),
	})
	// User 3 (visible supervisor) paid; item charged to user 1 -> entry for user 3.
	createReceiptWithItems(t, "Supervisor paid", 15, 3, 1, []commands.UpsertItemCommand{
		chargedItem("sup item", 15, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	if _, exists := result[2]; exists {
		t.Errorf("invisible counterparty (user 2) should be excluded from settlement, got %v", result)
	}
	assertOwed(t, result, 3, 15)
}

func TestGetAmountOwedForUserUnrestrictedViewerUnaffectedByIsolation(t *testing.T) {
	defer tearDownUserTest()
	setupIsolatedAmountOwedTest(t)

	// Elevate the caller to an admin (app.users.read) -> unrestricted visibility,
	// so isolation must not filter the settlement map.
	grantAppPerms(t, 1, permissions.AppUsersRead)

	createReceiptWithItems(t, "Peer paid", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("peer item", 10, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10)
}

// Cross-group settlement: a counterparty shared via an OPEN group still appears (for
// that group's portion), while their contribution from an ISOLATED group where the
// caller cannot see them is excluded. "Isolated means isolated" — the isolated portion
// is dropped even though the caller knows the counterparty from the open group.
func TestGetAmountOwedForUserCrossGroupExcludesIsolatedPortionKeepsOpen(t *testing.T) {
	defer tearDownUserTest()
	setupIsolatedAmountOwedTest(t) // group 1 isolated (sup=3); user 1 is the restricted caller

	// User 1 and user 2 ALSO share the OPEN group 2; user 1 may read both groups.
	grantGroupPerms(t, 1, 2, permissions.GroupReceiptsRead)
	grantGroupPerms(t, 2, 2, permissions.GroupReceiptsRead)

	// Isolated group 1: user 2 (invisible peer here) paid, item charged to user 1 ($10).
	r1 := createReceiptWithItems(t, "Iso peer paid", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("iso item", 10, 1),
	})
	// Open group 2: user 2 (visible here) paid, item charged to user 1 ($15).
	r2 := createReceiptWithItems(t, "Open peer paid", 15, 2, 2, []commands.UpsertItemCommand{
		chargedItem("open item", 15, 1),
	})

	w, result := callGetAmountOwed(1, "", []string{
		utils.UintToString(r1.ID), utils.UintToString(r2.ID),
	})
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	// Only the open-group portion counts; the isolated-group portion is excluded.
	assertOwed(t, result, 2, 15)
}

// --- A. Authorization ---------------------------------------------------

func TestGetAmountOwedForUserReturnsForbiddenWhenCallerNotInGroup(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 4 is only in Group 2; calling with groupId=1 must be rejected.
	w, _ := callGetAmountOwed(4, "1", nil)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

// --- B. Empty / baseline ------------------------------------------------

func TestGetAmountOwedForUserEmptyResultWhenNoReceiptsExist(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	w, result := callGetAmountOwed(1, "1", nil)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}
	if len(result) != 0 {
		t.Errorf("expected empty result map, got %v", result)
	}
}

func TestGetAmountOwedForUserExcludesItemsChargedToPayer(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Self-charged", 10, 1, 1, []commands.UpsertItemCommand{
		chargedItem("only item", 10, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}
	if len(result) != 0 {
		t.Errorf("self-charged items should be excluded; got %v", result)
	}
}

// --- C. Basic positive (single receipt) ---------------------------------

func TestGetAmountOwedForUserCallerChargedOnOthersReceipt(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 2 paid; item charged to user 1.
	createReceiptWithItems(t, "Lunch", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("burger", 10, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10) // caller owes user 2 $10
}

func TestGetAmountOwedForUserCallerPaidForOthersItem(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 1 paid; item charged to user 2.
	createReceiptWithItems(t, "Lunch", 10, 1, 1, []commands.UpsertItemCommand{
		chargedItem("burger", 10, 2),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, -10) // user 2 owes caller $10
}

// --- D. Multi-user / multi-receipt aggregation --------------------------

func TestGetAmountOwedForUserMultipleItemsSameUserSum(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Groceries", 30, 2, 1, []commands.UpsertItemCommand{
		chargedItem("apples", 10, 1),
		chargedItem("bread", 10, 1),
		chargedItem("milk", 10, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 30) // caller owes user 2 $30 in total
}

func TestGetAmountOwedForUserItemsChargedToMultipleUsersExcludesSelf(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Dinner", 30, 1, 1, []commands.UpsertItemCommand{
		chargedItem("steak (user 2)", 10, 2),
		chargedItem("salad (user 3)", 10, 3),
		chargedItem("dessert (user 1, self)", 10, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, -10)
	assertOwed(t, result, 3, -10)
	if _, exists := result[1]; exists {
		t.Errorf("self-charged item should not appear in result map, got %v", result)
	}
}

func TestGetAmountOwedForUserNetCancellationAcrossTwoReceipts(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 2 paid $10, item charged to user 1 → caller owes 10.
	createReceiptWithItems(t, "Lunch", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("sandwich", 10, 1),
	})
	// User 1 paid $10, item charged to user 2 → caller is owed 10. Net 0.
	createReceiptWithItems(t, "Coffee", 10, 1, 1, []commands.UpsertItemCommand{
		chargedItem("latte", 10, 2),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 0) // entry should exist and net to zero
}

// --- E. Negative (refund) coverage --------------------------------------

func TestGetAmountOwedForUserNegativeReceiptCallerCharged(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 2 received a $50 refund; item -$50 charged to user 1.
	// Semantics: user 2 owes the refund share back to user 1 → negative entry.
	createReceiptWithItems(t, "Store return", -50, 2, 1, []commands.UpsertItemCommand{
		chargedItem("returned shirt", -50, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, -50)
}

func TestGetAmountOwedForUserNegativeReceiptCallerPaid(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 1 received a $50 refund; item -$50 charged to user 2.
	// Semantics: caller must pass user 2's share of the refund → positive entry.
	createReceiptWithItems(t, "Store return", -50, 1, 1, []commands.UpsertItemCommand{
		chargedItem("returned shirt", -50, 2),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 50)
}

func TestGetAmountOwedForUserRefundCancelsOriginalDebt(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// Original purchase: user 2 paid $50, item charged to user 1 → caller owes 50.
	createReceiptWithItems(t, "Original", 50, 2, 1, []commands.UpsertItemCommand{
		chargedItem("widget", 50, 1),
	})
	// Refund: user 2 received refund -$50, item -$50 charged to user 1 → net 0.
	createReceiptWithItems(t, "Refund", -50, 2, 1, []commands.UpsertItemCommand{
		chargedItem("widget refund", -50, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 0)
}

func TestGetAmountOwedForUserMixedSignItemsInOneReceipt(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	// User 2 paid; one $20 item charged to user 1 and one -$20 item charged to user 1.
	// Net contribution to user 1's debt to user 2 is zero.
	createReceiptWithItems(t, "Mixed adjustments", 0, 2, 1, []commands.UpsertItemCommand{
		chargedItem("charge", 20, 1),
		chargedItem("adjustment", -20, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 0)
}

// --- F. Zero amount -----------------------------------------------------

func TestGetAmountOwedForUserZeroAmountItemContributesZero(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Free sample", 0, 2, 1, []commands.UpsertItemCommand{
		chargedItem("free item", 0, 1),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	// Zero may appear as either no entry or an entry of 0; both are acceptable.
	if got, ok := result[2]; ok && !got.Equal(decimal.Zero) {
		t.Errorf("zero-amount item should contribute zero; got resultMap[2]=%s", got.String())
	}
}

// --- G. Status filter ---------------------------------------------------

func TestGetAmountOwedForUserExcludesResolvedItems(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Mixed statuses", 20, 2, 1, []commands.UpsertItemCommand{
		chargedItemWithStatus("counted", 10, 1, models.ITEM_OPEN),
		chargedItemWithStatus("excluded", 10, 1, models.ITEM_RESOLVED),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10) // RESOLVED item is excluded
}

func TestGetAmountOwedForUserExcludesDraftItems(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	createReceiptWithItems(t, "Draft mix", 20, 2, 1, []commands.UpsertItemCommand{
		chargedItemWithStatus("counted", 10, 1, models.ITEM_OPEN),
		chargedItemWithStatus("excluded", 10, 1, models.ITEM_DRAFT),
	})

	w, result := callGetAmountOwed(1, "1", nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10) // DRAFT item is excluded
}

// --- H. All-group expansion --------------------------------------------

func TestGetAmountOwedForUserAllGroupAggregatesAcrossMemberships(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	db := repositories.GetDB()
	// Make user 1 a member of Group 2 as well so the all-group covers both groups.
	db.Create(&models.GroupMember{GroupID: 2, UserID: 1})
	// The all-group view now aggregates only groups the caller may actually read,
	// so grant receipts.read in group 2 as well (group 1 is granted by the fixture).
	grantGroupPerms(t, 1, 2, permissions.GroupReceiptsRead)

	// CreateAllGroup makes a new group with IsAllGroup=true and adds user 1 as OWNER member.
	groupRepository := repositories.NewGroupRepository(nil)
	allGroup, err := groupRepository.CreateAllGroup(1)
	if err != nil {
		t.Fatalf("failed to create all-group: %v", err)
	}
	grantGroupPerms(t, 1, allGroup.ID, permissions.GroupReceiptsRead)

	// Group 1: user 2 paid $10, item charged to user 1 → caller owes user 2.
	createReceiptWithItems(t, "G1 receipt", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("g1 item", 10, 1),
	})
	// Group 2: user 4 paid $25, item charged to user 1 → caller owes user 4.
	createReceiptWithItems(t, "G2 receipt", 25, 4, 2, []commands.UpsertItemCommand{
		chargedItem("g2 item", 25, 1),
	})

	w, result := callGetAmountOwed(1, strconv.FormatUint(uint64(allGroup.ID), 10), nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10)
	assertOwed(t, result, 4, 25)
}

// --- I. receiptIds parameter -------------------------------------------

func TestGetAmountOwedForUserReceiptIdsWithoutGroupId(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	receipt := createReceiptWithItems(t, "Lunch", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("burger", 10, 1),
	})

	// Empty groupId — handler skips group-role check and uses receiptIds directly.
	w, result := callGetAmountOwed(1, "", []string{strconv.FormatUint(uint64(receipt.ID), 10)})
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10)
}

func TestGetAmountOwedForUserReceiptIdsCombinedWithGroupId(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	db := repositories.GetDB()
	db.Create(&models.GroupMember{GroupID: 2, UserID: 1})
	// The out-of-group receipt is in group 2; authorize the caller there too.
	grantGroupPerms(t, 1, 2, permissions.GroupReceiptsRead)

	// In-group receipt (groupId path).
	createReceiptWithItems(t, "G1 receipt", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("g1 item", 10, 1),
	})
	// Out-of-group receipt (referenced explicitly via receiptIds).
	g2Receipt := createReceiptWithItems(t, "G2 receipt", 25, 4, 2, []commands.UpsertItemCommand{
		chargedItem("g2 item", 25, 1),
	})

	w, result := callGetAmountOwed(1, "1", []string{strconv.FormatUint(uint64(g2Receipt.ID), 10)})
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10)
	assertOwed(t, result, 4, 25)
}

func TestShouldNotAllowUserToGetPagedUsers(t *testing.T) {
	defer tearDownUserTest()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	GetPagedUsers(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotGetPagedUsersWithBadRequest(t *testing.T) {
	defer tearDownUserTest()

	tests := map[string]struct {
		input  commands.PagedRequestCommand
		expect int
	}{
		"badOrderBy": {
			input:  commands.PagedRequestCommand{Page: 1, PageSize: 50, OrderBy: "badOrderBy", SortDirection: "asc"},
			expect: http.StatusInternalServerError,
		},
		"badSortDirection": {
			input:  commands.PagedRequestCommand{Page: 1, PageSize: 50, OrderBy: "username", SortDirection: "badSortDirection"},
			expect: http.StatusBadRequest,
		},
		"badPage": {
			input:  commands.PagedRequestCommand{Page: -1, PageSize: 50, OrderBy: "username", SortDirection: "asc"},
			expect: http.StatusBadRequest,
		},
		"badPageSize": {
			input:  commands.PagedRequestCommand{Page: 1, PageSize: -2, OrderBy: "username", SortDirection: "asc"},
			expect: http.StatusBadRequest,
		},
		"valid": {
			input:  commands.PagedRequestCommand{Page: 1, PageSize: 25, OrderBy: "username", SortDirection: "asc"},
			expect: http.StatusOK,
		},
	}

	grantAllAppPerms(t, 1)

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
		r = r.WithContext(newContext)

		GetPagedUsers(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, name+" status "+strconv.Itoa(w.Result().StatusCode), test.expect)
		}
	}
}

func TestShouldAllowAdminToGetPagedUsers(t *testing.T) {
	defer tearDownUserTest()

	// grantAllAppPerms creates user 1 with the admin role; add two more so the
	// page returns a known, non-trivial set.
	grantAllAppPerms(t, 1)
	db := repositories.GetDB()
	db.Create(&models.User{Username: "alpha", DisplayName: "alpha", Password: "password"})
	db.Create(&models.User{Username: "beta", DisplayName: "beta", Password: "password"})

	command := commands.PagedRequestCommand{Page: 1, PageSize: 25, OrderBy: "username", SortDirection: "asc"}
	bytes, _ := json.Marshal(command)
	reader := strings.NewReader(string(bytes))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1}})
	r = r.WithContext(newContext)

	GetPagedUsers(w, r)

	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	var pagedData structs.PagedData
	if err := json.NewDecoder(w.Result().Body).Decode(&pagedData); err != nil {
		utils.PrintTestError(t, err, "no error decoding paged data")
		return
	}

	if pagedData.TotalCount != 3 {
		utils.PrintTestError(t, pagedData.TotalCount, int64(3))
	}
	if len(pagedData.Data) != 3 {
		utils.PrintTestError(t, len(pagedData.Data), 3)
	}
}

// TestResetPasswordClearsTheOidcProvisionedFlag pins the wiring behind the
// lockout guard's own advice.
//
// An account created by an OIDC provisioning login holds only a random,
// discarded password, so unlinking its last provider is refused with "ask an
// administrator to set one before disconnecting". That advice was a dead end:
// the administrator could comply and the unlink stayed refused, because nothing
// observed the reset. This is the observation.
func TestResetPasswordClearsTheOidcProvisionedFlag(t *testing.T) {
	defer tearDownUserTest()

	db := repositories.GetDB()

	provisioned := models.User{Username: "provisioned-user", Password: "unusable"}
	if err := db.Create(&provisioned).Error; err != nil {
		t.Fatalf("failed to create the user: %v", err)
	}

	provider := models.OidcProvider{
		Name:        "resetidp",
		DisplayName: "Reset IdP",
		IssuerUrl:   "https://idp.example.com",
		ClientId:    "client",
		Scope:       "openid",
	}
	if err := db.Create(&provider).Error; err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	identity := models.OidcIdentity{
		Subject:         "reset-subject",
		OidcProviderId:  provider.ID,
		UserId:          provisioned.ID,
		ProvisionedUser: true,
	}
	if err := db.Create(&identity).Error; err != nil {
		t.Fatalf("failed to create the identity: %v", err)
	}

	admin := uint(9001)
	grantAppPerms(t, admin, permissions.AppUsersUpdate)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/user/"+strconv.Itoa(int(provisioned.ID))+"/resetPassword", strings.NewReader(""))

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", strconv.Itoa(int(provisioned.ID)))
	ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chiContext)
	ctx = context.WithValue(ctx, "reset_password", structs.ResetPasswordCommand{Password: "a-real-password"})
	ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{
		CustomClaims: &structs.Claims{UserId: admin},
	})

	ResetPassword(recorder, request.WithContext(ctx))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var refreshed models.OidcIdentity
	if err := db.First(&refreshed, identity.ID).Error; err != nil {
		t.Fatalf("failed to reload the identity: %v", err)
	}

	if refreshed.ProvisionedUser {
		t.Error("setting a password must clear the provisioned flag, or the unlink stays refused forever")
	}
}

// The all-group amount-owed view must NOT fold in a group the caller cannot read.
// A member of a group with no receipts.read role there must not have that group's
// settlement leak through "All". Regression guard for the All-group read gate.
func TestGetAmountOwedForUserAllGroupExcludesUnreadableGroup(t *testing.T) {
	defer tearDownUserTest()
	setupAmountOwedTest(t)

	db := repositories.GetDB()
	// Member of group 2, but with NO group role (so no receipts.read there).
	db.Create(&models.GroupMember{GroupID: 2, UserID: 1})

	groupRepository := repositories.NewGroupRepository(nil)
	allGroup, err := groupRepository.CreateAllGroup(1)
	if err != nil {
		t.Fatalf("failed to create all-group: %v", err)
	}
	grantGroupPerms(t, 1, allGroup.ID, permissions.GroupReceiptsRead)

	// Group 1 (readable via fixture): caller owes user 2 $10.
	createReceiptWithItems(t, "G1 receipt", 10, 2, 1, []commands.UpsertItemCommand{
		chargedItem("g1 item", 10, 1),
	})
	// Group 2 (NOT readable): caller would owe user 4 $25 — must be excluded.
	createReceiptWithItems(t, "G2 receipt", 25, 4, 2, []commands.UpsertItemCommand{
		chargedItem("g2 item", 25, 1),
	})

	w, result := callGetAmountOwed(1, strconv.FormatUint(uint64(allGroup.ID), 10), nil)
	if w.Result().StatusCode != http.StatusOK {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusOK)
		return
	}

	assertOwed(t, result, 2, 10)
	// User 4's group-2 charge must NOT appear (group 2 is not readable).
	if _, ok := result[4]; ok {
		utils.PrintTestError(t, "group 2 (unreadable) leaked into all-group settlement", "excluded")
	}
}
