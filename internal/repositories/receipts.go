package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm/clause"
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

func GetReceiptGroupIdByReceiptId(id string) (uint, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("group_id").Find(&receipt).Error
	if err != nil {
		return 0, err
	}

	return receipt.GroupId, nil
}

func GetFullyLoadedReceiptById(id string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}
