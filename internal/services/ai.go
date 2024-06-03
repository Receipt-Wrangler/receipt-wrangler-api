package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"
)

func NewAiService(receiptProcessingSettingsId string) (*AiService, error) {
	repository := repositories.NewReceiptProcessingSettings(nil)
	client := &AiService{}

	receiptProcessingSettings, err := repository.GetReceiptProcessingSettingsById(receiptProcessingSettingsId)
	if err != nil {
		return nil, err
	}
	client.ReceiptProcessingSettings = receiptProcessingSettings

	return client, nil
}

type AiService struct {
	ReceiptProcessingSettings models.ReceiptProcessingSettings
}

// TODO: V5 create system task for full raw response
func (service *AiService) CreateChatCompletion(messages []structs.AiClientMessage, decryptKey bool) (string, error) {
	switch service.ReceiptProcessingSettings.AiType {

	case models.OPEN_AI_CUSTOM_NEW:
		return service.OpenAiCustomChatCompletion(messages)

	case models.OLLAMA:
		return service.OllamaChatCompletion(messages)

	case models.OPEN_AI_NEW:
		return service.OpenAiChatCompletion(messages, decryptKey)

	case models.GEMINI_NEW:
		return service.GeminiChatCompletion(messages, decryptKey)
	}

	return "", nil
}

func (service *AiService) OpenAiChatCompletion(messages []structs.AiClientMessage, decryptKey bool) (string, error) {
	key, err := service.getKey(decryptKey)
	if err != nil {
		return "", err
	}
	client := openai.NewClient(key)

	openAiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for index, message := range messages {
		openAiMessages[index] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    openAiMessages,
			N:           1,
			Temperature: 0,
		},
	)
	if err != nil {
		return "", err
	}

	response := resp.Choices[0].Message.Content
	return response, nil
}

func (service *AiService) GeminiChatCompletion(messages []structs.AiClientMessage, decryptKey bool) (string, error) {
	ctx := context.Background()
	key, err := service.getKey(decryptKey)
	if err != nil {
		return "", err
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	prompt := ""
	for _, aiMessage := range messages {
		prompt += aiMessage.Content + " "
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			json := fmt.Sprintf("%s", part)
			return json, nil
		}
	}

	return "", nil
}

func (service *AiService) OpenAiCustomChatCompletion(messages []structs.AiClientMessage) (string, error) {
	result := ""
	body := map[string]interface{}{
		"model":       service.ReceiptProcessingSettings.Model,
		"messages":    messages,
		"temperature": 0,
		"max_tokens":  -1,
		"stream":      false,
	}
	httpClient := http.Client{}
	httpClient.Timeout = 10 * time.Minute

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest(http.MethodPost, service.ReceiptProcessingSettings.Url, bodyBytesBuffer)
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Close = true

	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var responseObject structs.OpenAiCustomResponse

	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return "", err
	}

	if len(responseObject.Choices) >= 0 {
		result = responseObject.Choices[0].Message.Content
	}

	return result, nil
}

func (service *AiService) OllamaChatCompletion(messages []structs.AiClientMessage) (string, error) {
	prompt := messages[0].Content

	result := ""
	body := map[string]interface{}{
		"model":       service.ReceiptProcessingSettings.Model,
		"prompt":      prompt,
		"temperature": 0,
		"stream":      false,
	}
	httpClient := http.Client{}
	httpClient.Timeout = 10 * time.Minute

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)
	url := fmt.Sprintf("%s/api/generate", service.ReceiptProcessingSettings.Url)

	request, err := http.NewRequest(http.MethodPost, url, bodyBytesBuffer)
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Close = true

	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var responseObject structs.OllamaResponse
	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return "", err
	}

	if len(responseObject.Response) >= 0 {
		result = responseObject.Response
	}

	return result, nil
}

func (service *AiService) CheckConnectivity(ranByUserId uint, decryptKey bool) (models.SystemTask, error) {
	messages := []structs.AiClientMessage{
		{
			Role:    "user",
			Content: "Respond with 'hello' if you are there!",
		},
	}

	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:        models.RECEIPT_PROCESSING_SETTINGS_CONNECTIVITY_CHECK,
		RanByUserId: &ranByUserId,
	}

	startedAt := time.Now()
	response, err := service.CreateChatCompletion(messages, decryptKey)
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
	} else {
		systemTaskCommand.Status = models.SYSTEM_TASK_SUCCEEDED
		systemTaskCommand.ResultDescription = fmt.Sprintf(
			"The configured model responded with: %s in response to: %s",
			response, messages[0].Content)
	}
	endedAt := time.Now()

	systemTaskCommand.StartedAt = startedAt
	systemTaskCommand.EndedAt = &endedAt

	if service.ReceiptProcessingSettings.ID > 0 {
		systemTaskCommand.AssociatedEntityId = service.ReceiptProcessingSettings.ID
		systemTaskCommand.AssociatedEntityType = models.RECEIPT_PROCESSING_SETTINGS

		systemTaskRepository := repositories.NewSystemTaskRepository(nil)
		return systemTaskRepository.CreateSystemTask(systemTaskCommand)
	}

	return models.SystemTask{
		Status: systemTaskCommand.Status,
	}, nil
}

func (service *AiService) getKey(decryptKey bool) (string, error) {
	if decryptKey {
		return service.decryptKey()
	}

	return service.ReceiptProcessingSettings.Key, nil
}

func (service *AiService) decryptKey() (string, error) {
	return utils.DecryptB64EncodedData(config.GetEncryptionKey(), service.ReceiptProcessingSettings.Key)
}
