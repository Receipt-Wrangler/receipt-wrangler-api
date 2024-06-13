package commands

type ReceiptProcessingResult struct {
	Receipt                         UpsertReceiptCommand
	RawResponse                     string
	ChatCompletionSystemTaskCommand UpsertSystemTaskCommand
	PromptSystemTaskCommand         UpsertSystemTaskCommand
	OcrSystemTaskCommand            UpsertSystemTaskCommand
}
