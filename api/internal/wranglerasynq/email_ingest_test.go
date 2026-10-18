package wranglerasynq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"runtime"
	"testing"
)

// testApiRoot returns the absolute path of the api/ directory.
func testApiRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// file = .../api/internal/wranglerasynq/email_ingest_test.go
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// newMutableEmailServer returns an httptest.Server that serves whatever *body
// points at, evaluated at request time. This lets a test set the response body
// after seeding so it can reference the real seeded ids.
func newMutableEmailServer(t *testing.T, body *string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(*body))
	}))
	t.Cleanup(server.Close)
	return server
}

// ollamaEmailResponse wraps receiptJSON inside an Ollama chat-completion
// envelope, JSON-encoding the inner string properly.
func ollamaEmailResponse(t *testing.T, receiptJSON string) string {
	t.Helper()
	envelope := map[string]any{
		"model":      "test",
		"created_at": "2024-01-01T00:00:00Z",
		"message":    map[string]any{"role": "assistant", "content": receiptJSON},
		"done":       true,
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal ollama email envelope: %v", err)
	}
	return string(b)
}

// seedEmailIngestPipeline builds the minimum graph needed to run
// handleEmailProcessPayload:
//   - A User + a Group (non-All), with the user as a member.
//   - A Prompt + ReceiptProcessingSettings (Ollama pointed at serverURL).
//   - A SystemSettings row (ID=1) linking to those processing settings.
//   - A SystemEmail + GroupSettings with email defaults set.
//
// Returns the loaded GroupSettings, the seeded User, and the seeded Group.
func seedEmailIngestPipeline(t *testing.T, serverURL string) (models.GroupSettings, models.User, models.Group) {
	t.Helper()
	t.Setenv("BASE_PATH", testApiRoot())
	db := repositories.GetDB()

	user := models.User{Username: "email-ingest-user", Password: "p", DisplayName: "x"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	group := models.Group{Name: "email-ingest-group"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}

	member := models.GroupMember{GroupID: group.ID, UserID: user.ID}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed group member: %v", err)
	}

	prompt := models.Prompt{Name: "email-ingest-prompt", Prompt: "Extract: @emailBody"}
	if err := db.Create(&prompt).Error; err != nil {
		t.Fatalf("seed prompt: %v", err)
	}

	settings := models.ReceiptProcessingSettings{
		Name:     "email-ingest-rps",
		AiType:   models.OLLAMA,
		Url:      serverURL,
		Model:    "m",
		PromptId: prompt.ID,
	}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatalf("seed receipt processing settings: %v", err)
	}

	sysSettings := models.SystemSettings{
		BaseModel:                   models.BaseModel{ID: 1},
		ReceiptProcessingSettingsId: &settings.ID,
	}
	if err := db.Create(&sysSettings).Error; err != nil {
		t.Fatalf("seed system settings: %v", err)
	}

	systemEmail := models.SystemEmail{
		Host:     "localhost",
		Port:     "993",
		Username: "test@example.com",
		Password: "x",
	}
	if err := db.Create(&systemEmail).Error; err != nil {
		t.Fatalf("seed system email: %v", err)
	}

	groupSettings := models.GroupSettings{
		GroupId:                     group.ID,
		EmailDefaultReceiptStatus:   models.OPEN,
		EmailDefaultReceiptPaidById: &user.ID,
		SystemEmailId:               &systemEmail.ID,
	}
	if err := db.Create(&groupSettings).Error; err != nil {
		t.Fatalf("seed group settings: %v", err)
	}

	// Reload via repository so clause.Associations (incl. SystemEmail) are preloaded.
	loaded, err := repositories.NewGroupSettingsRepository(nil).GetGroupSettingsById(
		utils.UintToString(groupSettings.ID),
	)
	if err != nil {
		t.Fatalf("reload group settings: %v", err)
	}

	return loaded, user, group
}

// createEmailIngestTestTags seeds the same two tags as the quick-scan ingest
// tests ("tag-a", "tag-b") so tests can reference them by predictable id.
func createEmailIngestTestTags() {
	db := repositories.GetDB()
	db.Create(&models.Tag{Name: "tag-a"})
	db.Create(&models.Tag{Name: "tag-b"})
}

// runEmailIngest seeds the full graph, sets the mock AI body to an Ollama
// envelope wrapping aiReceiptJSON, and drives handleEmailProcessPayload with a
// text-only email (no attachment). Returns the fully-loaded created receipt.
func runEmailIngest(t *testing.T, aiReceiptJSON string) (models.Receipt, error) {
	t.Helper()

	var body string
	server := newMutableEmailServer(t, &body)
	groupSettings, _, _ := seedEmailIngestPipeline(t, server.URL)

	// Seed categories (ids 1-3 named "test"/"test2"/"test3") and tags.
	repositories.CreateTestCategories()
	createEmailIngestTestTags()

	// Set the body after seeding so it can reference the real db ids.
	body = ollamaEmailResponse(t, aiReceiptJSON)

	payload := EmailProcessTaskPayload{
		GroupSettingsId: groupSettings.ID,
		Metadata: structs.EmailMetadata{
			Subject: "Test Receipt",
			Body:    "Test email body text",
		},
	}

	if err := handleEmailProcessPayload("email-ingest-test-task", payload); err != nil {
		return models.Receipt{}, err
	}

	var created models.Receipt
	db := repositories.GetDB()
	if err := db.Where("created_by_string = ?", "Email Integration").First(&created).Error; err != nil {
		return models.Receipt{}, err
	}

	return repositories.NewReceiptRepository(nil).GetFullyLoadedReceiptById(utils.UintToString(created.ID))
}

func hasCategoryInReceipt(categories []models.Category, id uint) bool {
	for _, c := range categories {
		if c.ID == id {
			return true
		}
	}
	return false
}

// TestEmailIngest_ResolvesIdOnlyAiCategory mirrors
// TestQuickScan_ResolvesIdOnlyAiCategory for the email path: the AI returns a
// category and tag by id only (the shape the default prompt produces), and the
// handler must resolve the name from the DB before Validate() is called, or the
// receipt fails with "Name is required" — the production bug this test pins.
func TestEmailIngest_ResolvesIdOnlyAiCategory(t *testing.T) {
	defer repositories.TruncateTestDb()

	receipt, err := runEmailIngest(t,
		`{"name":"Walmart","amount":98.21,"date":"2024-01-01T00:00:00Z","categories":[{"id":1}],"tags":[{"id":1}]}`,
	)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	if len(receipt.Categories) != 1 || !hasCategoryInReceipt(receipt.Categories, 1) {
		utils.PrintTestError(t, receipt.Categories, "category 1 resolved")
		return
	}
	if receipt.Categories[0].Name != "test" {
		utils.PrintTestError(t, receipt.Categories[0].Name, "test")
	}

	if len(receipt.Tags) != 1 {
		utils.PrintTestError(t, receipt.Tags, "tag 1 resolved")
		return
	}
	if receipt.Tags[0].Name != "tag-a" {
		utils.PrintTestError(t, receipt.Tags[0].Name, "tag-a")
	}
}

// TestEmailIngest_DropsUnresolvableAiCategory mirrors
// TestQuickScan_DropsUnresolvableAiCategory: a hallucinated/deleted category id
// is dropped rather than failing the whole ingest.
func TestEmailIngest_DropsUnresolvableAiCategory(t *testing.T) {
	defer repositories.TruncateTestDb()

	receipt, err := runEmailIngest(t,
		`{"name":"Junk","amount":5.00,"date":"2024-01-01T00:00:00Z","categories":[{"id":1},{"id":999}]}`,
	)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	if len(receipt.Categories) != 1 || !hasCategoryInReceipt(receipt.Categories, 1) {
		utils.PrintTestError(t, receipt.Categories, "only category 1 (999 dropped)")
	}
}

