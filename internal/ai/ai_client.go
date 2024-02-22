package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/structs"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
)

func NewAiClient(clientType structs.AiClientType, openAiClient *openai.Client, geminiClient *genai.Client) *AiClient {
	return &AiClient{
		ClientType:   clientType,
		OpenAiClient: openAiClient,
		GeminiClient: geminiClient,
	}
}

type AiClient struct {
	ClientType   structs.AiClientType      `json:"clientType"`
	Messages     []structs.AiClientMessage `json:"messages"`
	OpenAiClient *openai.Client            `json:"openAiClient"`
	GeminiClient *genai.Client             `json:"geminiClient"`
}

func (aiClient *AiClient) CreateChatCompletion() (string, error) {
	switch aiClient.ClientType {

	case structs.LLAMA_GPT:
		return aiClient.LlamaGptChatCompletion()

	case structs.OPEN_AI:
		return aiClient.OpenAiChatCompletion()

	case structs.GEMINI:
		return aiClient.GeminiChatCompletion()
	}

	return "", nil
}

func (aiClient *AiClient) LlamaGptChatCompletion() (string, error) {
	result := ""
	config := config.GetConfig()
	body := map[string]interface{}{
		"messages":    aiClient.Messages,
		"temperature": 0,
	}
	httpClient := http.Client{}
	httpClient.Timeout = 10 * time.Minute

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	request, err := http.NewRequest(http.MethodPost, config.AiSettings.Url, bodyBytesBuffer)
	if err != nil {
		return "", err
	}

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

	var responseObject LlamaGptResponse

	err = json.Unmarshal(responseBody, &responseObject)
	if err != nil {
		return "", err
	}

	result = responseObject.Choices[0].Message.Content

	return result, nil
}

func (aiClient *AiClient) OpenAiChatCompletion() (string, error) {
	openAiMessages := make([]openai.ChatCompletionMessage, len(aiClient.Messages))
	for index, message := range aiClient.Messages {
		openAiMessages[index] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	resp, err := aiClient.OpenAiClient.CreateChatCompletion(
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

func (aiClient *AiClient) GeminiChatCompletion() (string, error) {
	ctx := context.Background()
	model := aiClient.GeminiClient.GenerativeModel("gemini-pro")
	prompt := ""
	for _, aiMessage := range aiClient.Messages {
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
