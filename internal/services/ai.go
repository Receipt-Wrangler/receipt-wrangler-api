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
	"receipt-wrangler/api/internal/constants"
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

func (service *AiService) CreateChatCompletion(options structs.AiChatCompletionOptions) (string, commands.UpsertSystemTaskCommand, error) {
	systemTask := commands.UpsertSystemTaskCommand{
		Type:                 models.CHAT_COMPLETION,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
		AssociatedEntityId:   service.ReceiptProcessingSettings.ID,
		StartedAt:            time.Now(),
	}
	response, rawResponse := "", ""
	err := error(nil)

	switch service.ReceiptProcessingSettings.AiType {
	case models.OPEN_AI_CUSTOM_NEW:
		response, rawResponse, err = service.OpenAiCustomChatCompletion(options)

	case models.OLLAMA:
		response, rawResponse, err = service.OllamaChatCompletion(options)

	case models.OPEN_AI_NEW:
		response, rawResponse, err = service.OpenAiChatCompletion(options)

	case models.GEMINI_NEW:
		response, rawResponse, err = service.GeminiChatCompletion(options)
	}
	if err != nil {
		endedAt := time.Now()
		systemTask.Status = models.SYSTEM_TASK_FAILED
		systemTask.ResultDescription = fmt.Sprintf("Error: %s; RawResponse: %s", err.Error(), rawResponse)
		systemTask.EndedAt = &endedAt

		return "", systemTask, err
	}

	endedAt := time.Now()
	systemTask.ResultDescription = rawResponse
	systemTask.EndedAt = &endedAt

	return response, systemTask, nil
}

func (service *AiService) OpenAiChatCompletion(options structs.AiChatCompletionOptions) (string, string, error) {
	key, err := service.getKey(options.DecryptKey)
	if err != nil {
		return "", "", err
	}
	client := openai.NewClient(key)

	openAiMessages := make([]openai.ChatCompletionMessage, len(options.Messages))

	if len(options.Messages) > 0 && len(options.Messages[0].Images) > 0 {
		for i, message := range options.Messages {
			chatParts := make([]openai.ChatMessagePart, 1+len(message.Images))

			chatParts[0] = openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: message.Content,
			}
			for j, image := range message.Images {
				imageUrl := openai.ChatMessageImageURL{
					URL:    image,
					Detail: openai.ImageURLDetailAuto,
				}

				imagePart := openai.ChatMessagePart{
					Type:     openai.ChatMessagePartTypeImageURL,
					ImageURL: &imageUrl,
				}

				chatParts[j+1] = imagePart
			}

			openAiMessages[i] = openai.ChatCompletionMessage{
				Role:         message.Role,
				MultiContent: chatParts,
			}
		}
	} else if len(options.Messages) > 0 {
		for index, message := range options.Messages {
			openAiMessages[index] = openai.ChatCompletionMessage{
				Role:    message.Role,
				Content: message.Content,
			}
		}
	}

	model := service.ReceiptProcessingSettings.Model
	if len(model) == 0 {
		model = openai.GPT3Dot5Turbo
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    openAiMessages,
			N:           1,
			Temperature: 0,
		},
	)
	if err != nil {
		responseBytes, _ := json.Marshal(resp)

		return "", string(responseBytes), err
	}

	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return "", "", err
	}

	response := resp.Choices[0].Message.Content
	return response, string(responseBytes), nil
}

func (service *AiService) GeminiChatCompletion(options structs.AiChatCompletionOptions) (string, string, error) {
	ctx := context.Background()
	key, err := service.getKey(options.DecryptKey)
	if err != nil {
		return "", "", err
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	prompt := ""
	for _, aiMessage := range options.Messages {
		prompt += aiMessage.Content + " "
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		responseBytes, _ := json.Marshal(resp)

		return "", string(responseBytes), err
	}

	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return "", "", err
	}

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {

			json := fmt.Sprintf("%s", part)
			return json, string(responseBytes), nil
		}
	}

	return "", "", nil
}

func (service *AiService) OpenAiCustomChatCompletion(options structs.AiChatCompletionOptions) (string, string, error) {
	result := ""
	body := map[string]interface{}{
		"model":       service.ReceiptProcessingSettings.Model,
		"messages":    options.Messages,
		"temperature": 0,
		"max_tokens":  -1,
		"stream":      false,
	}
	httpClient := http.Client{}
	httpClient.Timeout = constants.AiHttpTimeout

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest(http.MethodPost, service.ReceiptProcessingSettings.Url, bodyBytesBuffer)
	if err != nil {
		return "", "", err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Close = true

	response, err := httpClient.Do(request)
	if err != nil {
		responseBytes, _ := json.Marshal(response)

		return "", string(responseBytes), err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	var responseObject structs.OpenAiCustomResponse

	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return "", "", err
	}

	if len(responseObject.Choices) >= 0 {
		result = responseObject.Choices[0].Message.Content
	}

	responseBytes, err := json.Marshal(responseObject)
	if err != nil {
		return "", "", err
	}

	return result, string(responseBytes), nil
}

func (service *AiService) OllamaChatCompletion(options structs.AiChatCompletionOptions) (string, string, error) {
	result := ""
	body := map[string]interface{}{
		"model":       service.ReceiptProcessingSettings.Model,
		"messages":    options.Messages,
		"temperature": 0,
		"stream":      false,
	}
	httpClient := http.Client{}
	httpClient.Timeout = constants.AiHttpTimeout

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest(http.MethodPost, service.ReceiptProcessingSettings.Url, bodyBytesBuffer)
	if err != nil {
		return "", "", err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Close = true

	response, err := httpClient.Do(request)
	if err != nil {
		responseBytes, _ := json.Marshal(response)
		return "", string(responseBytes), err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	var responseObject structs.OllamaTextResponse
	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return "", "", err
	}

	if len(responseObject.Message.Content) >= 0 {
		result = responseObject.Message.Content
		result = utils.RemoveJsonFormat(result)
	}

	responseBytes, err := json.Marshal(responseObject)
	if err != nil {
		return "", "", err
	}

	return result, string(responseBytes), nil
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
	response, _, err := service.CreateChatCompletion(structs.AiChatCompletionOptions{
		Messages:   messages,
		DecryptKey: decryptKey,
	})
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
