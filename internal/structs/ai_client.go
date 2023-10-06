package structs

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/constants"
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

func (aiClient *AiClient) CreateChatCompletion() (string, error) {
	switch aiClient.ClientType {

	case LLAMA_GPT:
		aiClient.LlamaGptChatCompletion()

	case OPEN_AI:
		return "stub", nil
	}

	return "", nil
}

func (aiClient *AiClient) LlamaGptChatCompletion() (string, error) {
	body := map[string]interface{}{
		"messages": aiClient.Messages,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	bodyBytesBuffer := bytes.NewBuffer(bodyBytes)

	response, err := http.Post("", constants.APPLICATION_JSON, bodyBytesBuffer)
	if err != nil {
		return "", err
	}

	fmt.Println(response)

	return "hello", nil
}
