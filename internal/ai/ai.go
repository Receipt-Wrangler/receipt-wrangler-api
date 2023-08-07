package ai

import (
	"context"
	"encoding/json"
	"fmt"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
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
	categoriesString, err := getCategoriesString()
	if err != nil {
		return "", err
	}

	tagsString, err := getTagsString()
	if err != nil {
		return "", err
	}

	currentYear := simpleutils.UintToString(uint(time.Now().Year()))
	prompt := fmt.Sprintf(`
	Find the receipt's name, total cost, and date. Format as:
	{
		Name: store name,
		Amount: amount,
		Date: date in zulu,
		Categories: categories,
		Tags: tags
	}

	Omit any value if not found with confidence. Assume the date is in the year %s if not provided, and assume time values are empty. The amount must be a float or integer.

	Choose up to 2 categories from the given list based on the receipt's items and store name. If none fit, omit the result. Select only the id, like:
	{
		Id: category id
	}

	Emphasize the relationship between the category and the receipt.

	Categories: %s

	Follow the same process as described for categories for tags.

	Tags: %s

	Receipt text: %s
`, currentYear, categoriesString, tagsString, ocrText)

	return prompt, nil
}

func getCategoriesString() (string, error) {
	categoryRepository := repositories.NewCategoryRepository(nil)
	categories, err := categoryRepository.GetAllCategories("id, name")
	if err != nil {
		return "", err
	}

	categoriesBytes, err := json.Marshal(categories)
	if err != nil {
		return "", err
	}

	return string(categoriesBytes), nil
}

func getTagsString() (string, error) {
	tagsRepository := repositories.NewTagsRepository(nil)
	tags, err := tagsRepository.GetAllTags("id, name")
	if err != nil {
		return "", err
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return string(tagsBytes), nil
}
