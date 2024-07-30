package structs

type AiClientMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

type AiChatCompletionOptions struct {
	// Messages to send to the AI model
	Messages []AiClientMessage `json:"messages"`

	// Determines whether to decrypt the key
	DecryptKey bool `json:"decryptKey"`
}
