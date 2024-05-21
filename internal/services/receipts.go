package services

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mime/multipart"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"strconv"
	"time"
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
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)
	systemTaskService := NewSystemTaskService(nil)
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

	now := time.Now()

	receiptCommand, receiptProcessingMetadata, err := MagicFillFromImage(magicFillCommand)
	finishedAt := time.Now()

	var systemTask commands.UpsertSystemTaskCommand
	var fallbackSystemTask commands.UpsertSystemTaskCommand

	if receiptProcessingMetadata.ReceiptProcessingSettingsIdRan > 0 {

		systemTask = commands.UpsertSystemTaskCommand{
			Type:                 models.QUICK_SCAN,
			Status:               systemTaskService.BoolToSystemTaskStatus(receiptProcessingMetadata.DidReceiptProcessingSettingsSucceed),
			AssociatedEntityId:   receiptProcessingMetadata.ReceiptProcessingSettingsIdRan,
			AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
			StartedAt:            now,
			EndedAt:              &finishedAt,
			RanByUserId:          &token.UserId,
		}
	}

	if receiptProcessingMetadata.FallbackReceiptProcessingSettingsIdRan > 0 {
		fallbackSystemTask = commands.UpsertSystemTaskCommand{
			Type:                 models.QUICK_SCAN,
			Status:               systemTaskService.BoolToSystemTaskStatus(receiptProcessingMetadata.DidFallbackReceiptProcessingSettingsSucceed),
			AssociatedEntityId:   receiptProcessingMetadata.FallbackReceiptProcessingSettingsIdRan,
			AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
			StartedAt:            now,
			EndedAt:              &finishedAt,
			RanByUserId:          &token.UserId,
		}
	}

	if err != nil {
		systemTask.ResultDescription = receiptProcessingMetadata.RawResponse
		fallbackSystemTask.ResultDescription = receiptProcessingMetadata.FallbackRawResponse

		if receiptProcessingMetadata.ReceiptProcessingSettingsIdRan > 0 {
			_, err := systemTaskRepository.CreateSystemTask(systemTask)
			if err != nil {
				return models.Receipt{}, err
			}
		}

		if receiptProcessingMetadata.FallbackReceiptProcessingSettingsIdRan > 0 {
			_, err := systemTaskRepository.CreateSystemTask(fallbackSystemTask)
			if err != nil {
				return models.Receipt{}, err
			}
		}

		return models.Receipt{}, err
	}

	systemTask.ResultDescription = systemTaskService.BuildSuccessReceiptProcessResultDescription(receiptProcessingMetadata)

	receiptCommand.PaidByUserID = paidByUserId
	receiptCommand.Status = models.ReceiptStatus(status)
	receiptCommand.GroupId = groupId

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository.SetTransaction(tx)
		receiptImageRepository.SetTransaction(tx)
		systemTaskRepository.SetTransaction(tx)

		createdReceipt, err = receiptRepository.CreateReceipt(receiptCommand, token.UserId)
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

		if receiptProcessingMetadata.ReceiptProcessingSettingsIdRan > 0 {
			_, err = systemTaskRepository.CreateSystemTask(systemTask)
			if err != nil {
				return err
			}
		}

		if receiptProcessingMetadata.FallbackReceiptProcessingSettingsIdRan > 0 {
			_, err = systemTaskRepository.CreateSystemTask(fallbackSystemTask)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	return createdReceipt, nil
}
