package services

import (
	"os"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetReceiptByReceiptImageId(receiptImageId string) (models.Receipt, error) {
	db := db.GetDB()
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Select("receipt_id").First(&fileData).Error
	if err != nil {
		return models.Receipt{}, err
	}

	receipt, err := repositories.GetReceiptById(strconv.FormatUint(uint64(fileData.ReceiptId), 10))
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func DeleteReceipt(id string) error {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload("ImageFiles").Find(&receipt).Error
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err = tx.Select(clause.Associations).Delete(&receipt).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&receipt).Error
		if err != nil {
			return err
		}

		for _, f := range receipt.ImageFiles {
			path, _ := utils.BuildFilePath(utils.UintToString(f.ReceiptId), utils.UintToString(f.ID), f.Name)
			os.Remove(path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
