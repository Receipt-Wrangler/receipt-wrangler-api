package ai

import (
	config "receipt-wrangler/api/internal/env"

	"github.com/sashabaranov/go-openai"
)

var client *openai.Client

func InitOpenAIClient() {
	client = openai.NewClient(config.GetConfig().OpenAiKey)
}

func GetClient() *openai.Client {
	return client
}
