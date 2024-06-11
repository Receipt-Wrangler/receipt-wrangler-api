package commands

type ReceiptProcessingMetadata struct {
	ReceiptProcessingSettingsIdRan              uint
	DidReceiptProcessingSettingsSucceed         bool
	FallbackReceiptProcessingSettingsIdRan      uint
	DidFallbackReceiptProcessingSettingsSucceed bool
	RawResponse                                 string
	FallbackRawResponse                         string
	OcrSystemTaskCommand                        UpsertSystemTaskCommand
	ChatCompletionSystemTaskCommand             UpsertSystemTaskCommand
	FallbackOcrSystemTaskCommand                UpsertSystemTaskCommand
	FallbackChatCompletionSystemTaskCommand     UpsertSystemTaskCommand
}
