package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func tearDownReceiptProcessingSettings() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToGetReceiptProcessingSettings(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetPagedReceiptProcessingSettings(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotGetReceiptProcessingSettingsWithBadRequest(t *testing.T) {
	defer tearDownSystemEmailTest()

	tests := map[string]struct {
		input  commands.PagedRequestCommand
		expect int
	}{
		"badOrderBy": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      50,
				OrderBy:       "badOrderBy",
				SortDirection: "asc",
			},
			expect: http.StatusInternalServerError,
		},
		"badSortDirection": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      50,
				OrderBy:       "host",
				SortDirection: "badSortDirection",
			},
			expect: http.StatusBadRequest,
		},
		"badPage": {
			input: commands.PagedRequestCommand{
				Page:          -1,
				PageSize:      50,
				OrderBy:       "host",
				SortDirection: "asc",
			},
			expect: http.StatusBadRequest,
		},
		"badPageSize": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      -2,
				OrderBy:       "host",
				SortDirection: "asc",
			},
			expect: http.StatusBadRequest,
		},
		"good": {
			input: commands.PagedRequestCommand{
				Page:          1,
				PageSize:      50,
				OrderBy:       "name",
				SortDirection: "asc",
			},
			expect: http.StatusOK,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		GetPagedReceiptProcessingSettings(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}
	}
}

func TestShouldNotAllowUserToCreateReceiptProcessingSettings(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CreateReceiptProcessingSettings(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotCreateReceiptProcessingSettingsWithBadRequest(t *testing.T) {

	os.Setenv("ENCRYPTION_KEY", "test")

	tests := map[string]struct {
		input  commands.UpsertReceiptProcessingSettingsCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty command": {
			input:  commands.UpsertReceiptProcessingSettingsCommand{},
			expect: http.StatusBadRequest,
		},
		"missing name": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Description: "description",
				AiType:      models.GEMINI,
				OcrEngine:   models.TESSERACT,
				Key:         "key",
				PromptId:    1,
				NumWorkers:  1,
			},
			expect: http.StatusBadRequest,
		},
		"missing ocr engine": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "name",
				Description: "description",
				AiType:      models.GEMINI,
				Key:         "key",
				PromptId:    1,
				NumWorkers:  1,
			},
			expect: http.StatusBadRequest,
		},
		"valid gemini settings": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "Gemini",
				Description: "description",
				AiType:      models.GEMINI,
				OcrEngine:   models.TESSERACT,
				Key:         "key",
				PromptId:    1,
				NumWorkers:  1,
			},
			expect: http.StatusOK,
		},
		"valid openAi settings": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "OpenAi",
				Description: "description",
				AiType:      models.OPEN_AI,
				OcrEngine:   models.TESSERACT,
				Key:         "key",
				PromptId:    1,
				NumWorkers:  1,
			},
			expect: http.StatusOK,
		},
		"valid openAi custom settings": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "OpenAi Custom",
				Description: "description",
				AiType:      models.OPEN_AI_CUSTOM,
				OcrEngine:   models.EASY_OCR,
				Key:         "optionalKey",
				Url:         "http://localhost:8080/v1/chat/completions",
				PromptId:    1,
				NumWorkers:  1,
			},
			expect: http.StatusOK,
		},
	}

	for name, test := range tests {
		db := repositories.GetDB()
		db.Create(&models.Prompt{})

		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		CreateReceiptProcessingSettings(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}

		tearDownReceiptProcessingSettings()
	}
}
