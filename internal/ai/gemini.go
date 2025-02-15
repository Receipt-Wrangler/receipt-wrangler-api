package ai

import (
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type GeminiClient struct {
	BaseClient
}

func NewGeminiClient(
	options structs.AiChatCompletionOptions,
	receiptProcessingSettings models.ReceiptProcessingSettings,
) *GeminiClient {
	return &GeminiClient{
		BaseClient{
			Options:                   options,
			ReceiptProcessingSettings: receiptProcessingSettings,
		},
	}
}

func (gemini GeminiClient) GetChatCompletion() (structs.ChatCompletionResult, error) {
	result := structs.ChatCompletionResult{}

	ctx := context.Background()
	key, err := gemini.getKey(gemini.Options.DecryptKey)
	if err != nil {
		return result, err
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return result, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	prompt := ""
	for _, aiMessage := range gemini.Options.Messages {
		prompt += aiMessage.Content + " "
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		responseBytes, _ := json.Marshal(resp)
		result.RawResponse = string(responseBytes)
		return result, err
	}

	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return result, err
	}

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {

			json := fmt.Sprintf("%s", part)
			result.Response = json
			result.RawResponse = string(responseBytes)
			return result, err
		}
	}

	return result, nil
}
