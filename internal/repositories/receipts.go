package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func GetReceiptById(receiptId string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func GetReceiptsByGroupId(groupId string) ([]models.Receipt, error) {
	db := db.GetDB()
	var receipts []models.Receipt

	err := db.Model(models.Receipt{}).Where("group_id = ?", groupId).Preload("Tags").Preload("Categories").Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	return receipts, nil
}
