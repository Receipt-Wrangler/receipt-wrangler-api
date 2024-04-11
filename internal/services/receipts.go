package services

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"strconv"
	"time"

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

func beforeUpdateReceipt(tx *gorm.DB, currentReceipt models.Receipt, updatedReceipt models.Receipt) (err error) {
	if updatedReceipt.GroupId > 0 && currentReceipt.GroupId != updatedReceipt.GroupId && len(currentReceipt.ImageFiles) > 0 {
		var oldGroup models.Group
		var newGroup models.Group

		err = tx.Table("groups").Where("id = ?", currentReceipt.GroupId).Select("id", "name").Find(&oldGroup).Error
		if err != nil {
			return err
		}

		err = tx.Table("groups").Where("id = ?", updatedReceipt.GroupId).Select("id", "name").Find(&newGroup).Error
		if err != nil {
			return err
		}

		oldGroupPath, err := simpleutils.BuildGroupPathString(simpleutils.UintToString(oldGroup.ID), oldGroup.Name)
		if err != nil {
			return err
		}

		newGroupPath, err := simpleutils.BuildGroupPathString(simpleutils.UintToString(newGroup.ID), newGroup.Name)
		if err != nil {
			return err
		}

		for _, fileData := range currentReceipt.ImageFiles {
			filename := simpleutils.BuildFileName(simpleutils.UintToString(currentReceipt.ID), simpleutils.UintToString(fileData.ID), fileData.Name)

			oldFilePath := filepath.Join(oldGroupPath, filename)
			newFilePathPath := filepath.Join(newGroupPath, filename)

			err := os.Rename(oldFilePath, newFilePathPath)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func UpdateReceipt(id string, command commands.UpsertReceiptCommand) (models.Receipt, error) {
	db := repositories.GetDB()
	var currentReceipt models.Receipt

	updatedReceipt, err := command.ToReceipt()
	if err != nil {
		return models.Receipt{}, err
	}

	err = db.Table("receipts").Where("id = ?", id).Preload("ImageFiles").Find(&currentReceipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	// NOTE: ID and field used for afterReceiptUpdated
	updatedReceipt.ID = currentReceipt.ID
	updatedReceipt.ResolvedDate = currentReceipt.ResolvedDate

	err = db.Transaction(func(tx *gorm.DB) error {

		txErr := beforeUpdateReceipt(tx, currentReceipt, updatedReceipt)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&currentReceipt).Updates(&updatedReceipt).Error
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("Tags").Replace(&updatedReceipt.Tags)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("Categories").Replace(&updatedReceipt.Categories)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("ReceiptItems").Replace(&updatedReceipt.ReceiptItems)
		if txErr != nil {
			return txErr
		}

		err = afterReceiptUpdated(tx, &updatedReceipt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	return updatedReceipt, nil
}

func afterReceiptUpdated(tx *gorm.DB, updatedReceipt *models.Receipt) error {
	err := tx.Where("receipt_id IS NULL").Delete(&models.Item{}).Error
	if err != nil {
		return err
	}

	if updatedReceipt.ID > 0 && updatedReceipt.Status == models.RESOLVED && updatedReceipt.ResolvedDate == nil {
		now := time.Now().UTC()
		err = tx.Table("receipts").Where("id = ?", updatedReceipt.ID).Update("resolved_date", now).Error
	} else if updatedReceipt.ID > 0 && updatedReceipt.Status != models.RESOLVED && updatedReceipt.ResolvedDate != nil {
		err = tx.Table("receipts").Where("id = ?", updatedReceipt.ID).Update("resolved_date", nil).Error
	}
	if err != nil {
		return err
	}

	if updatedReceipt.Status == models.RESOLVED && updatedReceipt.ID > 0 {
		err := updateItemsToStatus(tx, updatedReceipt, models.ITEM_RESOLVED)
		if err != nil {
			return err
		}
	}

	if updatedReceipt.Status == models.DRAFT && updatedReceipt.ID > 0 {
		err := updateItemsToStatus(tx, updatedReceipt, models.ITEM_DRAFT)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateItemsToStatus(tx *gorm.DB, r *models.Receipt, status models.ItemStatus) error {
	var items []models.Item
	var itemIdsToUpdate []uint

	err := tx.Table("items").Where("receipt_id = ?", r.ID).Find(&items).Error
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Status != status {
			itemIdsToUpdate = append(itemIdsToUpdate, item.ID)
		}
	}

	if len(itemIdsToUpdate) > 0 {
		err := tx.Table("items").Where("id IN ?", itemIdsToUpdate).UpdateColumn("status", status).Error
		if err != nil {
			return err
		}
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

	receiptCommand, err := MagicFillFromImage(magicFillCommand)
	if err != nil {
		return models.Receipt{}, err
	}

	receiptCommand.PaidByUserID = paidByUserId
	receiptCommand.Status = models.ReceiptStatus(status)
	receiptCommand.GroupId = groupId

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository.SetTransaction(tx)
		receiptImageRepository.SetTransaction(tx)

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

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	return createdReceipt, nil
}
