package services

import (
	"mime/multipart"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
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
		var imagesToDelete []string
		fileRepository := repositories.NewFileRepository(nil)
		fileRepository.SetTransaction(tx)

		for _, f := range receipt.ImageFiles {
			path, _ := fileRepository.BuildFilePath(simpleutils.UintToString(f.ReceiptId), simpleutils.UintToString(f.ID), f.Name)
			imagesToDelete = append(imagesToDelete, path)
		}

		err = tx.Select(clause.Associations).Delete(&receipt).Error
		if err != nil {
			return err
		}

		for _, path := range imagesToDelete {
			os.Remove(path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func QuickScan(
	token *structs.Claims,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	paidByUserId uint,
	groupId uint,
	status models.ReceiptStatus,
) (models.Receipt, error) {
	db := repositories.GetDB()
	var createdReceipt models.Receipt

	fileRepository := repositories.NewFileRepository(nil)

	fileBytes := make([]byte, fileHeader.Size)
	_, err := file.Read(fileBytes)
	if err != nil {
		return models.Receipt{}, err
	}
	defer file.Close()

	validatedFileType, err := fileRepository.ValidateFileType(fileBytes)
	if err != nil {
		return models.Receipt{}, err
	}

	magicFillCommand := commands.MagicFillCommand{
		ImageData: fileBytes,
		Filename:  fileHeader.Filename,
	}

	receiptRepository := repositories.NewReceiptRepository(nil)
	receiptImageRepository := repositories.NewReceiptImageRepository(nil)

	receipt, err := MagicFillFromImage(magicFillCommand)
	if err != nil {
		return models.Receipt{}, err
	}

	receipt.PaidByUserID = paidByUserId
	receipt.Status = models.ReceiptStatus(status)
	receipt.GroupId = groupId

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository.SetTransaction(tx)
		receiptImageRepository.SetTransaction(tx)

		createdReceipt, err = receiptRepository.CreateReceipt(receipt, token.UserId)
		if err != nil {
			return err
		}

		fileData := models.FileData{
			Name:      fileHeader.Filename,
			Size:      uint(fileHeader.Size),
			ReceiptId: createdReceipt.ID,
			FileType:  validatedFileType,
		}

		_, err := receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	return createdReceipt, nil
}
