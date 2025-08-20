package ai

import (
	"encoding/json"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"strings"

	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/context"
)

type OpenAiClient struct {
	BaseClient
}

func NewOpenAiClient(
	options structs.AiChatCompletionOptions,
	receiptProcessingSettings models.ReceiptProcessingSettings,
) *OpenAiClient {
	return &OpenAiClient{
		BaseClient{
			Options:                   options,
			ReceiptProcessingSettings: receiptProcessingSettings,
		},
	}
}

func (openAi OpenAiClient) GetChatCompletion() (structs.ChatCompletionResult, error) {
	result := structs.ChatCompletionResult{}
	var config openai.ClientConfig

	key, err := openAi.getKey(openAi.Options.DecryptKey)
	if err != nil {
		return result, err
	}

	if strings.Contains(openAi.ReceiptProcessingSettings.Url, "azure") {
		config = openai.DefaultAzureConfig(key, openAi.ReceiptProcessingSettings.Url)
	} else {
		config = openai.DefaultConfig(key)
	}

	if len(openAi.ReceiptProcessingSettings.Url) > 0 {
		config.BaseURL = openAi.ReceiptProcessingSettings.Url
	}
	client := openai.NewClientWithConfig(config)

	openAiMessages := make([]openai.ChatCompletionMessage, len(openAi.Options.Messages))

	if len(openAi.Options.Messages) > 0 && len(openAi.Options.Messages[0].Images) > 0 {
		for i, message := range openAi.Options.Messages {
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
	} else if len(openAi.Options.Messages) > 0 {
		for index, message := range openAi.Options.Messages {
			openAiMessages[index] = openai.ChatCompletionMessage{
				Role:    message.Role,
				Content: message.Content,
			}
		}
	}

	model := openAi.ReceiptProcessingSettings.Model
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
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		responseBytes, _ := json.Marshal(resp)
		result.RawResponse = string(responseBytes)

		return result, err
	}

	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return result, err
	}

	result.RawResponse = string(responseBytes)
	result.Response = resp.Choices[0].Message.Content
	return result, nil
}
