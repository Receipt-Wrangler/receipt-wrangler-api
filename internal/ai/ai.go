package ai

import (
	"context"
	"encoding/json"
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"time"

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
	logger := logging.GetLogger()

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
	logger.Print(openAiResponse, "raw response")

	err = json.Unmarshal([]byte(openAiResponse), &result)
	if err != nil {
		return models.Receipt{}, err
	}

	return result, nil
}

func getPrompt(ocrText string) string {
	currentYear := simpleutils.UintToString(uint(time.Now().Year()))
	return fmt.Sprintf(`
	Find the name of the receipt, total cost of the receipt, and receipt date.
	Format the data in valid json as follows:
	
	{
		Name: store name here,
		Amount: receipt amount here,
		Date: receipt date here
	}
	
	If the data cannot be found with confidence, do not make up a value, omit the value from the results.
	The date must be a valid date, formatted in zulu time. If no year is provided, assume it is the year %s. Assume time values are empty.
	The amount must be a valid float, or integer.

	%s
	`, currentYear, ocrText)
}
