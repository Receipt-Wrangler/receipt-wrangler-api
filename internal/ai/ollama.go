package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type OllamaClient struct {
	BaseClient
}

func NewOllamaClient(
	options structs.AiChatCompletionOptions,
	receiptProcessingSettings models.ReceiptProcessingSettings,
) *OllamaClient {
	return &OllamaClient{
		BaseClient{
			Options:                   options,
			ReceiptProcessingSettings: receiptProcessingSettings,
		},
	}
}

func (ollama OllamaClient) GetChatCompletion() (structs.ChatCompletionResult, error) {
	result := structs.ChatCompletionResult{}
	body := map[string]interface{}{
		"model":       ollama.ReceiptProcessingSettings.Model,
		"messages":    ollama.Options.Messages,
		"temperature": 0,
		"stream":      false,
		"format":      "json",
	}
	httpClient := http.Client{}
	httpClient.Timeout = constants.AiHttpTimeout

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return result, err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest(http.MethodPost, ollama.ReceiptProcessingSettings.Url, bodyBytesBuffer)
	if err != nil {
		return result, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Close = true

	response, err := httpClient.Do(request)
	if err != nil {
		responseBytes, _ := json.Marshal(response)
		result.RawResponse = string(responseBytes)
		return result, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	var responseObject structs.OllamaTextResponse
	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return result, err
	}

	if len(responseObject.Message.Content) >= 0 {
		result.Response = responseObject.Message.Content
	}

	responseBytes, err := json.Marshal(responseObject)
	if err != nil {
		result.RawResponse = string(responseBytes)
		return result, err
	}
	result.RawResponse = string(responseBytes)

	return result, nil
}
