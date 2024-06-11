package models

import (
	"database/sql/driver"
	"errors"
)

type AiClientType string

const (
	OPEN_AI_CUSTOM     AiClientType = "openAiCustom"
	OPEN_AI            AiClientType = "openAi"
	GEMINI             AiClientType = "gemini"
	OPEN_AI_CUSTOM_NEW AiClientType = "OPEN_AI_CUSTOM"
	OPEN_AI_NEW        AiClientType = "OPEN_AI"
	GEMINI_NEW         AiClientType = "GEMINI"
	OLLAMA             AiClientType = "OLLAMA"
)

func (clientType *AiClientType) Scan(value string) error {
	*clientType = AiClientType(value)
	return nil
}

func (clientType AiClientType) Value() (driver.Value, error) {
	if len(clientType) == 0 {
		return "", nil
	}

	if clientType !=
		OPEN_AI_CUSTOM &&
		clientType != OPEN_AI &&
		clientType != GEMINI &&
		clientType != OPEN_AI_CUSTOM_NEW &&
		clientType != OPEN_AI_NEW &&
		clientType != GEMINI_NEW &&
		clientType != OLLAMA {
		return nil, errors.New("invalid ai client type")
	}
	return string(clientType), nil
}
