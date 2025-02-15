package ai

import (
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type BaseClient struct {
	Options                   structs.AiChatCompletionOptions
	ReceiptProcessingSettings models.ReceiptProcessingSettings
}

func (baseClient BaseClient) getKey(decryptKey bool) (string, error) {
	if decryptKey && len(baseClient.ReceiptProcessingSettings.Key) > 0 {
		return baseClient.decryptKey()
	}

	return baseClient.ReceiptProcessingSettings.Key, nil
}

func (baseClient BaseClient) decryptKey() (string, error) {
	return utils.DecryptB64EncodedData(config.GetEncryptionKey(), baseClient.ReceiptProcessingSettings.Key)
}
