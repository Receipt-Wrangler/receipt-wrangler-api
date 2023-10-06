package structs

import (
	"database/sql/driver"
	"errors"
)

type AiClientType string

const (
	LLAMA_GPT AiClientType = "llamaGpt"
	OPEN_AI   AiClientType = "openAi"
)

func (clientType *AiClientType) Scan(value string) error {
	*clientType = AiClientType(value)
	return nil
}

func (clientType AiClientType) Value() (driver.Value, error) {
	if clientType != LLAMA_GPT {
		return nil, errors.New("invalid ai client type")
	}
	return string(clientType), nil
}

type AiClientMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewAiClient(clientType AiClientType) *AiClient {
	return &AiClient{
		ClientType: clientType,
	}
}

type AiClient struct {
	ClientType AiClientType      `json:"clientType"`
	Messages   []AiClientMessage `json:"messages"`
}

func (aiClient *AiClient) CreateChatCompletion() {

}
