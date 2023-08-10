package repositories

import (
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type ReceiptImageRepository struct {
	BaseRepository
}

func NewReceiptImageRepository(tx *gorm.DB) ReceiptImageRepository {
	repository := ReceiptImageRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository ReceiptImageRepository) GetReceiptImageById(receiptImageId uint) (models.FileData, error) {
	db := repository.GetDB()
	var result models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Find(&result).Error
	if err != nil {
		return models.FileData{}, err
	}

	return result, nil
}
