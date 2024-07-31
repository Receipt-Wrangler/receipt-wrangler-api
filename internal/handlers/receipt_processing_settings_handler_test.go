package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"os"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
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

func TestShouldTestValidAndInvalidCreateReceiptProcessingSettingCommands(t *testing.T) {
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
			},
			expect: http.StatusOK,
		},
		"valid ollama settings": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:          "Ollama",
				Description:   "description",
				AiType:        models.OLLAMA,
				Model:         "llama3",
				IsVisionModel: true,
				OcrEngine:     models.EASY_OCR,
				Url:           "http://localhost:8080/v1/chat/completions",
				PromptId:      1,
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

func TestShouldNotAllowUserToGetReceiptProcessingSettingsById(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	GetReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotReceiptProcessingSettingsByIdDueToBadId(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/", reader)

	var expectedStatusCode = http.StatusInternalServerError

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldGetReceiptProcessingSettingsById(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/", reader)

	var expectedStatusCode = http.StatusOK

	db := repositories.GetDB()
	db.Create(&models.Prompt{})
	db.Create(&models.ReceiptProcessingSettings{
		PromptId: 1,
	})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	GetReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToUpdateReceiptProcessingSettingsById(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	UpdateReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotUpdateReceiptProcessingSettingsByIdDueToBadId(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/", reader)

	var expectedStatusCode = http.StatusInternalServerError

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	UpdateReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldTestValidAndInvalidUpdateReceiptProcessingSettingCommands(t *testing.T) {
	defer tearDownReceiptProcessingSettings()

	os.Setenv("ENCRYPTION_KEY", "test")
	db := repositories.GetDB()
	db.Create(&models.Prompt{
		Name: "prompt",
	})
	db.Create(&models.ReceiptProcessingSettings{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		PromptId: 1,
	})

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
			},
			expect: http.StatusBadRequest,
		},
		"valid ollama vision": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:          "name",
				Description:   "description",
				AiType:        models.OLLAMA,
				Model:         "llava",
				Url:           "http://localhost:8080/v1/chat/completions",
				IsVisionModel: true,
				PromptId:      1,
			},
			expect: http.StatusOK,
		},
		"valid ollama": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "name",
				Description: "description",
				AiType:      models.OLLAMA,
				OcrEngine:   models.EASY_OCR,
				Model:       "llama3",
				Url:         "http://localhost:8080/v1/chat/completions",
				PromptId:    1,
			},
			expect: http.StatusOK,
		},
		"valid gemini settings": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:        "Gemini",
				Description: "description",
				AiType:      models.GEMINI,
				OcrEngine:   models.TESSERACT,
				Key:         "key",
				PromptId:    1,
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
			},
			expect: http.StatusOK,
		},
		"valid openAi vision": {
			input: commands.UpsertReceiptProcessingSettingsCommand{
				Name:          "OpenAi",
				Description:   "description",
				AiType:        models.OPEN_AI,
				Model:         "gpt-4o",
				IsVisionModel: true,
				Key:           "key",
				PromptId:      1,
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
			},
			expect: http.StatusOK,
		},
	}

	for name, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		chiContext := chi.NewRouteContext()
		chiContext.URLParams.Add("id", "1")
		routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
		r = r.WithContext(routeContext)

		newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
		r = r.WithContext(newContext)

		UpdateReceiptProcessingSettingsById(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}
	}
}

func TestShouldNotAllowUserToDeleteReceiptProcessingSettingsById(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	DeleteReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotDeleteReceiptProcessingSettingsByIdDueToBadId(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/", reader)

	var expectedStatusCode = http.StatusInternalServerError

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "badId")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeleteReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldDeleteReceiptProcessingSettingsById(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	w := httptest.NewRecorder()
	reader := strings.NewReader("")
	r := httptest.NewRequest("POST", "/api/", reader)

	var expectedStatusCode = http.StatusOK

	db := repositories.GetDB()
	db.Create(&models.Prompt{})
	db.Create(&models.ReceiptProcessingSettings{
		PromptId: 1,
	})

	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("id", "1")
	routeContext := context.WithValue(r.Context(), chi.RouteCtxKey, chiContext)
	r = r.WithContext(routeContext)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.ADMIN}})
	r = r.WithContext(newContext)

	DeleteReceiptProcessingSettingsById(w, r)

	if w.Result().StatusCode != expectedStatusCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedStatusCode)
	}
}

func TestShouldNotAllowUserToCheckReceiptProcessingSettingsConnectivity(t *testing.T) {
	defer tearDownReceiptProcessingSettings()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	newContext := context.WithValue(r.Context(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{CustomClaims: &structs.Claims{UserId: 1, UserRole: models.USER}})
	r = r.WithContext(newContext)

	CheckReceiptProcessingSettingsConnectivity(w, r)

	if w.Result().StatusCode != http.StatusForbidden {
		utils.PrintTestError(t, w.Result().StatusCode, http.StatusForbidden)
	}
}

func TestShouldNotCheckReceiptProcessingSettingsConnectivityWithBadRequest(t *testing.T) {
	defer tearDownReceiptProcessingSettings()

	os.Setenv("ENCRYPTION_KEY", "test")

	key, _ := utils.EncryptAndEncodeToBase64(config.GetEncryptionKey(), "key")

	repositories.CreateTestUser()

	ocrEngine := models.TESSERACT

	db := repositories.GetDB()
	db.Create(&models.Prompt{})
	db.Create(&models.ReceiptProcessingSettings{
		Name:        "Gemini",
		Description: "description",
		AiType:      models.GEMINI,
		OcrEngine:   &ocrEngine,
		Key:         key,
		PromptId:    1,
	})

	tests := map[string]struct {
		input  commands.CheckReceiptProcessingSettingsCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
		"empty command": {
			input:  commands.CheckReceiptProcessingSettingsCommand{},
			expect: http.StatusBadRequest,
		},
		"empty id": {
			input: commands.CheckReceiptProcessingSettingsCommand{
				ID: 0,
			},
			expect: http.StatusBadRequest,
		},
		"empty command and id": {
			input: commands.CheckReceiptProcessingSettingsCommand{
				ID:                                     0,
				UpsertReceiptProcessingSettingsCommand: commands.UpsertReceiptProcessingSettingsCommand{},
			},
			expect: http.StatusBadRequest,
		},
		"badId": {
			input: commands.CheckReceiptProcessingSettingsCommand{
				ID: 200,
			},
			expect: http.StatusInternalServerError,
		},
		"goodId": {
			input: commands.CheckReceiptProcessingSettingsCommand{
				ID:                                     1,
				UpsertReceiptProcessingSettingsCommand: commands.UpsertReceiptProcessingSettingsCommand{},
			},
			expect: http.StatusOK,
		},
		"good command": {
			input: commands.CheckReceiptProcessingSettingsCommand{
				ID: 0,
				UpsertReceiptProcessingSettingsCommand: commands.UpsertReceiptProcessingSettingsCommand{
					Name:        "Gemini",
					Description: "description",
					AiType:      models.GEMINI,
					OcrEngine:   models.TESSERACT,
					Key:         "key",
					PromptId:    1,
				},
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

		CheckReceiptProcessingSettingsConnectivity(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, fmt.Sprintf("%s expected: %d", name, test.expect))
		}
	}
}
