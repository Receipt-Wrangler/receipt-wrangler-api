package repositories

import (
	"gorm.io/gorm"
)

type ReceiptProcessingSettings struct {
	BaseRepository
}

func NewReceiptProcessingSettings(tx *gorm.DB) ReceiptProcessingSettings {
	repository := ReceiptProcessingSettings{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}
