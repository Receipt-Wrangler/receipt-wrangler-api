package ai

import (
	"encoding/json"
	"fmt"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
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

	model := client.GenerativeModel(gemini.ReceiptProcessingSettings.Model)
	model.GenerationConfig.ResponseMIMEType = "application/json"
	parts := make([]genai.Part, 0)
	for _, aiMessage := range gemini.Options.Messages {
		parts = append(parts, genai.Text(aiMessage.Content))

		if len(aiMessage.Images) > 0 {
			for _, image := range aiMessage.Images {
				imageBytes, err := utils.Base64Decode(image)
				if err != nil {
					return result, err
				}

				blob := genai.Blob{
					MIMEType: utils.GetMimeType(imageBytes).String(),
					Data:     imageBytes,
				}

				parts = append(parts, blob)
			}
		}
	}

	resp, err := model.GenerateContent(ctx, parts...)
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
			partText := fmt.Sprintf("%s", part)
			result.Response = partText
			result.RawResponse = string(responseBytes)
			return result, err
		}
	}

	return result, nil
}
