package ai

type OpenAiCustomMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAiCustomChoice struct {
	Index   int                 `json:"index"`
	Message OpenAiCustomMessage `json:"message"`
}

type OpenAiCustomResponse struct {
	Choices []OpenAiCustomChoice `json:"choices"`
}
