package ai

import (
	"context"
	"encoding/json"
	"fmt"
	db "receipt-wrangler/api/internal/database"
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
	prompt, err := getPrompt(ocrText)
	if err != nil {
		return models.Receipt{}, err
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
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

func getPrompt(ocrText string) (string, error) {
	var categories []models.Category
	db := db.GetDB()
	db.Model(models.Category{}).Select("id", "name").Find(&categories)

	categoriesBytes, err := json.Marshal(categories)
	if err != nil {
		return "", err
	}

	categoriesString := string(categoriesBytes)

	currentYear := simpleutils.UintToString(uint(time.Now().Year()))
	prompt := fmt.Sprintf(`
	Find the name of the receipt, total cost of the receipt, and receipt date.
	Format the data in valid json as follows:
	
	{
		Name: store name here,
		Amount: receipt amount here,
		Date: receipt date here
		Categories: categories here
	}
	
	If the data cannot be found with confidence, do not make up a value, omit the value from the results.
	The date must be a valid date, formatted in zulu time. If no year is provided, assume it is the year %s. Assume time values are empty.
	The amount must be a valid float, or integer.

	Next, I will give you a list of categories. Based on the data found, choose a maximum of 2 categories that best categorises the receipt based on the items of the receipt, and the receipt's store name. If none are suitable, please omit the result.
	Use the name of the category to make your selections.
	Select only the id, and format the results as follows:
	{
		Id: id of category here
	}

	It is not required to select the maximum number of categories, but try to emphasize the relationship between the category and the receipt based on the data found.

	Categories to choose from: 
	%s

	Receipt text:
	%s
	`, currentYear, categoriesString, ocrText)

	return prompt, nil
}
