package ai

type LlamaGptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlamaGptChoice struct {
	Index   int             `json:"index"`
	Message LlamaGptMessage `json:"message"`
}

type LlamaGptResponse struct {
	Choices []LlamaGptChoice `json:"choices"`
}
