package structs

type ReceiptProcessingMetadata struct {
	ReceiptProcessingSettingsIdRan              uint
	DidReceiptProcessingSettingsSucceed         bool
	FallbackReceiptProcessingSettingsIdRan      uint
	DidFallbackReceiptProcessingSettingsSucceed bool
	RawResponse                                 string
	FallbackRawResponse                         string
}
