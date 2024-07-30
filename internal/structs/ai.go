package structs

type AiClientMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiChatCompletionOptions struct {
	// Messages to send to the AI model
	Messages []AiClientMessage `json:"messages"`

	// Image path used for vision models
	ImagePath string `json:"imagePath"`

	// Determines whether to decrypt the key
	DecryptKey bool `json:"decryptKey"`
}
