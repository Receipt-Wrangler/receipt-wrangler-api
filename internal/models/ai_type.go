package models

import (
	"database/sql/driver"
	"errors"
)

type AiClientType string

const (
	OPEN_AI_CUSTOM AiClientType = "openAiCustom"
	OPEN_AI        AiClientType = "openAi"
	GEMINI         AiClientType = "gemini"
)

func (clientType *AiClientType) Scan(value string) error {
	*clientType = AiClientType(value)
	return nil
}

func (clientType AiClientType) Value() (driver.Value, error) {
	if len(clientType) == 0 {
		return "", nil
	}

	if clientType != OPEN_AI_CUSTOM && clientType != OPEN_AI && clientType != GEMINI {
		return nil, errors.New("invalid ai client type")
	}
	return string(clientType), nil
}
