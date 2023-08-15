package services

import (
	"os"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetReceiptByReceiptImageId(receiptImageId string) (models.Receipt, error) {
	db := repositories.GetDB()
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Select("receipt_id").First(&fileData).Error
	if err != nil {
		return models.Receipt{}, err
	}

	receiptRepository := repositories.NewReceiptRepository(nil)
	receipt, err := receiptRepository.GetReceiptById(strconv.FormatUint(uint64(fileData.ReceiptId), 10))
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func DeleteReceipt(id string) error {
	db := repositories.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload("ImageFiles").Find(&receipt).Error
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		fileRepository := repositories.NewFileRepository(nil)
		fileRepository.SetTransaction(tx)

		err = tx.Select(clause.Associations).Delete(&receipt).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&receipt).Error
		if err != nil {
			return err
		}

		for _, f := range receipt.ImageFiles {
			path, _ := fileRepository.BuildFilePath(simpleutils.UintToString(f.ReceiptId), simpleutils.UintToString(f.ID), f.Name)
			os.Remove(path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
