package ai

import (
	"context"
	"encoding/json"
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"

	"github.com/sashabaranov/go-openai"
)

var client *openai.Client

func InitOpenAIClient() {
	client = openai.NewClient(config.GetConfig().OpenAiKey)
}

func GetClient() *openai.Client {
	return client
}

func ReadReceiptData(ocrText string) (models.Receipt, error) {
	var result models.Receipt

	client := GetClient()
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: getPrompt(ocrText),
				},
			},
			N:           1,
			Temperature: 0,
		},
	)
	if err != nil {
		return models.Receipt{}, err
	}

	openAiResponse := resp.Choices[0].Message.Content
	err = json.Unmarshal([]byte(openAiResponse), &result)
	if err != nil {
		return models.Receipt{}, nil
	}

	return result, nil
}

func getPrompt(ocrText string) string {
	return fmt.Sprintf(`Could you find the total cost of this receipt, the name of the store, and the date of the receipt?
	Please respond with nothing else other than the data.
	Please format the data as follows:
	
	{
		name: store name here,
		amount: receipt amount here,
		date: receipt date here
	}
	
	If these values cannot be found confidently, please enter "null" instead of making up a value.
	The date must be a valid date.
	The amount must be a valid float, or integer.
	The data must be valid JSON.

	%s
	`, ocrText)
}