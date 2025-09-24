package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

func uintPtr(u uint) *uint {
	return &u
}

func setupReceiptsTest() {
	repositories.CreateTestGroupWithUsers()
	repositories.CreateTestCategories()
}

func tearDownReceiptsTest() {
	repositories.TruncateTestDb()
}

func TestShouldGetPagedReceiptsWithFullReceipts(t *testing.T) {
	defer tearDownReceiptsTest()
	setupReceiptsTest()

	// Create a test receipt with receipt items and comments
	receiptRepository := repositories.NewReceiptRepository(nil)
	testCommand := commands.UpsertReceiptCommand{
		Name:         "Test Receipt",
		Amount:       decimal.NewFromFloat(10.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		GroupId:      1,
		Status:       models.OPEN,
		Items: []commands.UpsertItemCommand{
			{
				Name:   "Test Item",
				Amount: decimal.NewFromFloat(10.00),
				Status: models.ITEM_OPEN,
			},
		},
		Comments: []commands.UpsertCommentCommand{
			{
				Comment: "Test comment",
				UserId:  uintPtr(1),
			},
		},
	}

	_, err := receiptRepository.CreateReceipt(testCommand, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Test with fullReceipts = true
	requestBody := commands.ReceiptPagedRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:     1,
			PageSize: 10,
		},
		FullReceipts: true,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	reader := strings.NewReader(string(body))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/receipt/group/1", reader)

	// Add path parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("groupId", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Add JWT context
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetPagedReceiptsForGroup(w, r)

	var pagedData structs.PagedData
	err = json.Unmarshal(w.Body.Bytes(), &pagedData)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
		return
	}

	if len(pagedData.Data) == 0 {
		utils.PrintTestError(t, "No receipts returned", "At least 1 receipt expected")
		return
	}

	// Check that receipts have full associations loaded
	receiptData, _ := json.Marshal(pagedData.Data[0])
	var receipt models.Receipt
	json.Unmarshal(receiptData, &receipt)

	if len(receipt.ReceiptItems) == 0 {
		utils.PrintTestError(t, "ReceiptItems not loaded", "ReceiptItems should be loaded with fullReceipts=true")
		return
	}

	if len(receipt.Comments) == 0 {
		utils.PrintTestError(t, "Comments not loaded", "Comments should be loaded with fullReceipts=true")
		return
	}

	// Clean up - no need to explicitly delete as tearDownReceiptsTest() handles it
}

func TestShouldGetPagedReceiptsWithoutFullReceipts(t *testing.T) {
	defer tearDownReceiptsTest()
	setupReceiptsTest()

	// Create a test receipt
	receiptRepository := repositories.NewReceiptRepository(nil)
	testCommand := commands.UpsertReceiptCommand{
		Name:         "Test Receipt",
		Amount:       decimal.NewFromFloat(10.00),
		Date:         time.Now(),
		PaidByUserID: 1,
		GroupId:      1,
		Status:       models.OPEN,
	}

	_, err := receiptRepository.CreateReceipt(testCommand, 1, true)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Test with fullReceipts = false (or not set)
	requestBody := commands.ReceiptPagedRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:     1,
			PageSize: 10,
		},
		FullReceipts: false,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	reader := strings.NewReader(string(body))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/receipt/group/1", reader)

	// Add path parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("groupId", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Add JWT context
	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetPagedReceiptsForGroup(w, r)

	var pagedData structs.PagedData
	err = json.Unmarshal(w.Body.Bytes(), &pagedData)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if w.Result().StatusCode != 200 {
		utils.PrintTestError(t, w.Result().StatusCode, 200)
		return
	}

	if len(pagedData.Data) == 0 {
		utils.PrintTestError(t, "No receipts returned", "At least 1 receipt expected")
		return
	}

	// Check that receipts have basic data but no full associations
	receiptData, _ := json.Marshal(pagedData.Data[0])
	var receipt models.Receipt
	json.Unmarshal(receiptData, &receipt)

	// Basic fields should be present
	if receipt.Name == "" {
		utils.PrintTestError(t, "Name not loaded", "Basic receipt fields should be loaded")
		return
	}

	// Clean up - no need to explicitly delete as tearDownReceiptsTest() handles it
}

func TestReceiptPagedRequestCommandShouldParseFullReceipts(t *testing.T) {
	// Test that FullReceipts field is correctly parsed from JSON
	jsonWithFullReceipts := `{
		"page": 1,
		"pageSize": 10,
		"fullReceipts": true,
		"filter": {}
	}`

	var command commands.ReceiptPagedRequestCommand
	err := json.Unmarshal([]byte(jsonWithFullReceipts), &command)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if !command.FullReceipts {
		utils.PrintTestError(t, command.FullReceipts, true)
		return
	}

	// Test default value (should be false)
	jsonWithoutFullReceipts := `{
		"page": 1,
		"pageSize": 10,
		"filter": {}
	}`

	var commandDefault commands.ReceiptPagedRequestCommand
	err = json.Unmarshal([]byte(jsonWithoutFullReceipts), &commandDefault)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if commandDefault.FullReceipts {
		utils.PrintTestError(t, commandDefault.FullReceipts, false)
		return
	}
}
