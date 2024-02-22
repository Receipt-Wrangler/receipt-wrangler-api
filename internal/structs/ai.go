package structs

import (
	"database/sql/driver"
	"errors"
)

type AiClientType string

const (
	LLAMA_GPT AiClientType = "llamaGpt"
	OPEN_AI   AiClientType = "openAi"
	GEMINI    AiClientType = "gemini"
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
