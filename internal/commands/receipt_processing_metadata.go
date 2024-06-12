package commands

type ReceiptProcessingMetadata struct {
	ReceiptProcessingSettingsIdRan              uint
	DidReceiptProcessingSettingsSucceed         bool
	FallbackReceiptProcessingSettingsIdRan      uint
	DidFallbackReceiptProcessingSettingsSucceed bool
	RawResponse                                 string
	FallbackRawResponse                         string
	Prompt                                      string
	PromptId                                    uint
	FallbackPrompt                              string
	FallbackPromptId                            uint
	OcrSystemTaskCommand                        UpsertSystemTaskCommand
	ChatCompletionSystemTaskCommand             UpsertSystemTaskCommand
	FallbackOcrSystemTaskCommand                UpsertSystemTaskCommand
	FallbackChatCompletionSystemTaskCommand     UpsertSystemTaskCommand
}
