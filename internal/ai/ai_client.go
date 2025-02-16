package ai

import (
	"receipt-wrangler/api/internal/structs"
)

type Client interface {
	GetChatCompletion() (structs.ChatCompletionResult, error)
}
