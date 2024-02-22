package services

import (
	"context"
	"encoding/json"
	"fmt"
	"receipt-wrangler/api/internal/ai"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"time"

	"google.golang.org/api/option"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
)

var client *openai.Client
var geminiClient *genai.Client

func InitOpenAIClient() error {
	config := config.GetConfig()
	apiKey := config.AiSettings.Key
	if len(apiKey) == 0 && config.AiSettings.AiType == structs.OPEN_AI {
		apiKey = config.AiSettings.Key
	}

	if len(apiKey) == 0 {
		return fmt.Errorf("OpenAI API key not found")
	}

	client = openai.NewClient(apiKey)
	return nil
}

func GetClient() *openai.Client {
	return client
}

func InitGeminiClient() error {
	ctx := context.Background()
	config := config.GetConfig()

	if len(config.AiSettings.Key) == 0 {
		return fmt.Errorf("Gemini API key not found")
	}

	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AiSettings.Key))
	if err != nil {
		return err
	}
	geminiClient = client

	return nil
}

func GetGeminiClient() *genai.Client {
	return geminiClient
}

func ReadReceiptData(ocrText string) (models.Receipt, error) {
	var result models.Receipt
	logger := logging.GetLogger()
	config := config.GetConfig()
	client := GetClient()
	geminiClient := GetGeminiClient()

	aiType := config.AiSettings.AiType
	if len(aiType) == 0 {
		aiType = structs.OPEN_AI
	}

	aiClient := ai.NewAiClient(aiType, client, geminiClient)
	clientMessages := []structs.AiClientMessage{}

	prompt, err := getPrompt(ocrText)
	if err != nil {
		return models.Receipt{}, err
	}

	clientMessages = append(clientMessages, structs.AiClientMessage{
		Role:    "user",
		Content: prompt,
	})
	aiClient.Messages = clientMessages

	response, err := aiClient.CreateChatCompletion()
	if err != nil {
		return models.Receipt{}, err
	}

	logger.Print(response, "raw response")

	err = json.Unmarshal([]byte(response), &result)
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
		"name": store name,
		"amount": amount,
		"date": date in zulu, with ALL time values set to 0,
		"categories": categories,
		"tags": tags
	}
	If a store name cannot be confidently found, use 'Default store name' as the default name.
	Omit any value if not found with confidence. Assume the date is in the year %s if not provided.
	The amount must be a float or integer.

	Please do NOT add any additional information, only valid JSON.
	Please return the json in plaintext ONLY, do not ever return it in a code block or any other format.

	Choose up to 2 categories from the given list based on the receipt's items and store name. If no categories fit, please return an empty array for the field and do not select any categories. When selecting categories, select only the id, like:
	{
		Id: category id
	}

	Emphasize the relationship between the category and the receipt, and use the description of the category to fine tune the results. Do not return categories that have an empty name or do not exist.


	Categories: %s

	Follow the same process as described for categories for tags.

	Tags: %s

	Receipt text: %s
`, currentYear, categoriesString, tagsString, ocrText)

	return prompt, nil
}

func getCategoriesString() (string, error) {
	categoryRepository := repositories.NewCategoryRepository(nil)
	categories, err := categoryRepository.GetAllCategories("id, name, description")
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
