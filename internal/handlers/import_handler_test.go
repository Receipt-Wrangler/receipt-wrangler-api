package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"
	"strings"
	"testing"
)

func tearDownImportTests() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToImportConfig(t *testing.T) {
	defer tearDownImportTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	ImportConfigJson(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldRunHandlerWithDifferentInputs(t *testing.T) {
	defer tearDownSystemEmailTest()

	os.Setenv("ENCRYPTION_KEY", "test")

	tests := map[string]struct {
		input                                      structs.Config
		expect                                     int
		expectReceiptProcessingSettingsToBeCreated bool
		expectSystemEmailsToBeCreated              bool
	}{
		"emptyBody": {
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: false,
		},
		"emptyConfig": {
			input:  structs.Config{},
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: false,
		},
		"valid config with no Ai settings": {
			input: structs.Config{
				SecretKey:            "doesn't get imported",
				EmailPollingInterval: 1200,
				EmailSettings: []structs.EmailSettings{
					structs.EmailSettings{
						Host:     "test",
						Port:     123,
						Username: "yeet",
						Password: "yah",
					},
				},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: true,
					AiPoweredReceipts: true,
				},
				Debug: structs.DebugConfig{
					DebugOcr: true,
				},
			},
			expect:                        http.StatusOK,
			expectSystemEmailsToBeCreated: true,
		},
		"invalid config with incomplete AI settings": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType: models.OPEN_AI,
				},
				EmailPollingInterval: 1200,
				EmailSettings: []structs.EmailSettings{
					structs.EmailSettings{
						Host:     "test",
						Port:     123,
						Username: "yeet",
					},
				},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: true,
					AiPoweredReceipts: true,
				},
				Debug: structs.DebugConfig{
					DebugOcr: true,
				},
			},
			expect: http.StatusInternalServerError,
		},
		"invalid config with incomplete email settings": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType:    models.OPEN_AI,
					Key:       "test",
					OcrEngine: models.TESSERACT,
				},
				EmailPollingInterval: 1200,
				EmailSettings: []structs.EmailSettings{
					structs.EmailSettings{
						Host:     "test",
						Port:     123,
						Username: "yeet",
					},
				},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: true,
					AiPoweredReceipts: true,
				},
				Debug: structs.DebugConfig{
					DebugOcr: true,
				},
			},
			expect: http.StatusInternalServerError,
		},
		"valid open AI config": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType:    models.OPEN_AI,
					Key:       "test",
					OcrEngine: models.TESSERACT,
				},
				EmailPollingInterval: 1200,
				EmailSettings: []structs.EmailSettings{
					structs.EmailSettings{
						Host:     "test",
						Port:     123,
						Username: "yeet",
						Password: "yah",
					},
				},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: true,
					AiPoweredReceipts: true,
				},
				Debug: structs.DebugConfig{
					DebugOcr: true,
				},
			},
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: true,
			expectSystemEmailsToBeCreated:              true,
		},
		"valid Gemini config": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType:     models.GEMINI,
					Key:        "test",
					OcrEngine:  models.TESSERACT,
					NumWorkers: 11,
				},
				EmailPollingInterval: 1200,
				EmailSettings:        []structs.EmailSettings{},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: false,
					AiPoweredReceipts: false,
				},
				Debug: structs.DebugConfig{
					DebugOcr: false,
				},
			},
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: true,
			expectSystemEmailsToBeCreated:              true,
		},
		"valid Open AI Custom config": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType:     models.OPEN_AI_CUSTOM,
					Key:        "test",
					Url:        "test",
					OcrEngine:  models.TESSERACT,
					NumWorkers: 10,
				},
				EmailPollingInterval: 1200,
				EmailSettings:        []structs.EmailSettings{},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: false,
					AiPoweredReceipts: false,
				},
				Debug: structs.DebugConfig{
					DebugOcr: false,
				},
			},
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: true,
			expectSystemEmailsToBeCreated:              true,
		},
		"valid Open AI Custom config, with missing ocr engine": {
			input: structs.Config{
				SecretKey: "doesn't get imported",
				AiSettings: structs.AiSettings{
					AiType:     models.OPEN_AI_CUSTOM,
					Key:        "test",
					Url:        "test",
					NumWorkers: 10,
				},
				EmailPollingInterval: 1200,
				EmailSettings:        []structs.EmailSettings{},
				Features: structs.FeatureConfig{
					EnableLocalSignUp: false,
					AiPoweredReceipts: false,
				},
				Debug: structs.DebugConfig{
					DebugOcr: false,
				},
			},
			expect: http.StatusOK,
			expectReceiptProcessingSettingsToBeCreated: true,
			expectSystemEmailsToBeCreated:              true,
		},
	}

	for name, test := range tests {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		testFilePath := "./test.json"

		inputBytes, err := json.Marshal(test.input)
		if err != nil {
			t.Error(err)
			return
		}

		err = os.WriteFile(testFilePath, inputBytes, 0644)
		if err != nil {
			t.Error(err)
			return
		}

		file, err := os.Open(testFilePath)
		if err != nil {
			t.Error(err)
			return
		}
		defer file.Close()
		defer os.Remove(testFilePath)

		part, err := writer.CreateFormFile("file", filepath.Base(testFilePath))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = io.Copy(part, file)
		if err != nil {
			t.Error(err)
			return
		}

		err = writer.Close()
		if err != nil {
			t.Error(err)
			return
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", body)
		r.Header.Set("Content-Type", writer.FormDataContentType())

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		ImportConfigJson(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}

		if test.expectReceiptProcessingSettingsToBeCreated {
			promptRepository := repositories.NewPromptRepository(nil)
			repository := repositories.NewSystemSettingsRepository(nil)
			systemSettings, err := repository.GetSystemSettings()
			if err != nil {
				t.Error(err)
				return
			}

			if systemSettings.ReceiptProcessingSettings.ID == 0 {
				utils.PrintTestError(t,
					"ReceiptProcessingSettings do not exist",
					fmt.Sprintf("%s expected: %s", name, "ReceiptProcessingSettingsToExist"))
			}

			if strings.Count(systemSettings.ReceiptProcessingSettings.Name, "Imported Settings") == 0 {
				utils.PrintTestError(t, systemSettings.ReceiptProcessingSettings.Name, "Imported Settings")
			}

			if systemSettings.EnableLocalSignUp != test.input.Features.EnableLocalSignUp {
				utils.PrintTestError(t, systemSettings.EnableLocalSignUp, test.input.Features.EnableLocalSignUp)
			}

			if systemSettings.DebugOcr != test.input.Debug.DebugOcr {
				utils.PrintTestError(t, systemSettings.DebugOcr, test.input.Debug.DebugOcr)
			}

			if systemSettings.NumWorkers != test.input.AiSettings.NumWorkers &&
				systemSettings.NumWorkers != 1 {
				utils.PrintTestError(t, systemSettings.NumWorkers, test.input.AiSettings.NumWorkers)
			}

			if len(systemSettings.ReceiptProcessingSettings.Key) > 0 {
				cleartextKey, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemSettings.ReceiptProcessingSettings.Key)
				if err != nil {
					t.Error(err)
					return
				}

				if cleartextKey != test.input.AiSettings.Key {
					utils.PrintTestError(t, cleartextKey, test.input.AiSettings.Key)
				}
			}

			prompt, err := promptRepository.GetPromptById("1")
			if err != nil {
				utils.PrintTestError(t, err, nil)
			}

			if prompt.Name != "Default Prompt" {
				utils.PrintTestError(t, prompt.Name, "Default Prompt")
			}
		}

		if test.expectSystemEmailsToBeCreated {
			db := repositories.GetDB()
			for _, emailSetting := range test.input.EmailSettings {
				var systemEmail models.SystemEmail
				err := db.Model(&models.SystemEmail{}).Where("host = ?", emailSetting.Host).Find(&systemEmail).Error
				if err != nil {
					t.Error(err)
					return
				}

				if systemEmail.Host != emailSetting.Host {
					utils.PrintTestError(t, systemEmail.Host, emailSetting.Host)
				}

				if systemEmail.Port != strconv.Itoa(emailSetting.Port) {
					utils.PrintTestError(t, systemEmail.Port, emailSetting.Port)
				}

				if systemEmail.Username != emailSetting.Username {
					utils.PrintTestError(t, systemEmail.Username, emailSetting.Username)
				}

				cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
				if err != nil {
					t.Error(err)
					return
				}

				if cleartextPassword != emailSetting.Password {
					utils.PrintTestError(t, cleartextPassword, emailSetting.Password)
				}
			}
		}
	}
}
