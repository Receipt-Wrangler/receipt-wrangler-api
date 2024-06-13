package commands

type ReceiptProcessingMetadata struct {
	ReceiptProcessingSettingsIdRan              uint
	DidReceiptProcessingSettingsSucceed         bool
	FallbackReceiptProcessingSettingsIdRan      uint
	DidFallbackReceiptProcessingSettingsSucceed bool
	RawResponse                                 string
	FallbackRawResponse                         string
	PromptSystemTaskCommand                     UpsertSystemTaskCommand
	FallbackPromptSystemTaskCommand             UpsertSystemTaskCommand
	OcrSystemTaskCommand                        UpsertSystemTaskCommand
	ChatCompletionSystemTaskCommand             UpsertSystemTaskCommand
	FallbackOcrSystemTaskCommand                UpsertSystemTaskCommand
	FallbackChatCompletionSystemTaskCommand     UpsertSystemTaskCommand
}
