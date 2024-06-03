package services

import (
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"sync"
)

func ReadReceiptImage(receiptImageId string) (commands.UpsertReceiptCommand, commands.ReceiptProcessingMetadata, error) {
	var result commands.UpsertReceiptCommand
	var pathToReadFrom string
	systemReceiptProcessingService, err := NewSystemReceiptProcessingService(nil)
	if err != nil {
		return result, commands.ReceiptProcessingMetadata{}, err
	}

	receiptImageUint, err := simpleutils.StringToUint(receiptImageId)
	if err != nil {
		return result, commands.ReceiptProcessingMetadata{}, err
	}

	receiptImageRepository := repositories.NewReceiptImageRepository(nil)
	receiptImage, err := receiptImageRepository.GetReceiptImageById(receiptImageUint)
	if err != nil {
		return result, commands.ReceiptProcessingMetadata{}, err
	}
	fileRepository := repositories.NewReceiptImageRepository(nil)

	receiptImagePath, err := fileRepository.BuildFilePath(simpleutils.UintToString(receiptImage.ReceiptId), receiptImageId, receiptImage.Name)
	if err != nil {
		return result, commands.ReceiptProcessingMetadata{}, err
	}

	receiptImageBytes, err := utils.ReadFile(receiptImagePath)
	if err != nil {
		return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
	}

	// TODO: make generic
	if receiptImage.FileType == constants.APPLICATION_PDF {
		bytes, err := fileRepository.ConvertPdfToJpg(receiptImageBytes)
		if err != nil {
			return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
		}

		pathToReadFrom, err = fileRepository.WriteTempFile(bytes)
		if err != nil {
			return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
		}

		defer os.Remove(pathToReadFrom)
	} else {
		pathToReadFrom = receiptImagePath
	}

	return systemReceiptProcessingService.ReadReceiptImage(pathToReadFrom)
}

func ReadReceiptImageFromFileOnly(path string) (commands.UpsertReceiptCommand, commands.ReceiptProcessingMetadata, error) {
	receiptProcessingService, err := NewSystemReceiptProcessingService(nil)
	if err != nil {
		return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
	}

	return receiptProcessingService.ReadReceiptImage(path)
}

func MagicFillFromImage(command commands.MagicFillCommand) (commands.UpsertReceiptCommand, commands.ReceiptProcessingMetadata, error) {
	fileRepository := repositories.NewFileRepository(nil)
	receiptProcessingService, err := NewSystemReceiptProcessingService(nil)
	if err != nil {
		return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
	}

	bytes, err := fileRepository.GetBytesFromImageBytes(command.ImageData)
	if err != nil {
		return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
	}

	filePath, err := fileRepository.WriteTempFile(bytes)
	if err != nil {
		return commands.UpsertReceiptCommand{}, commands.ReceiptProcessingMetadata{}, err
	}
	defer os.Remove(filePath)

	return receiptProcessingService.ReadReceiptImage(filePath)
}

func GetReceiptImagesForGroup(groupId string, userId string) ([]models.FileData, error) {
	db := repositories.GetDB()
	groupRepository := repositories.NewGroupRepository(nil)
	groupService := NewGroupService(nil)
	groupIds := make([]uint, 0)

	group, err := groupRepository.GetGroupById(groupId, false)
	if err != nil {
		return nil, err
	}

	if group.IsAllGroup {
		groups, err := groupService.GetGroupsForUser(userId)
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			groupIds = append(groupIds, group.ID)
		}
	} else {
		uintGroupId, err := simpleutils.StringToUint(groupId)
		if err != nil {
			return nil, err
		}

		groupIds = append(groupIds, uintGroupId)
	}

	fileDataResults := make([]models.FileData, 0)
	err = db.Table("receipts").Select("receipts.id, receipts.group_id, file_data.*").Joins("inner join file_data on file_data.receipt_id=receipts.id").Where("receipts.group_id IN ?", groupIds).Scan(&fileDataResults).Error
	if err != nil {
		return nil, err
	}

	return fileDataResults, nil
}

func ReadAllReceiptImagesForGroup(groupId string, userId string) ([]structs.OcrExport, error) {
	fileRepository := repositories.NewFileRepository(nil)
	fileDataResults, err := GetReceiptImagesForGroup(groupId, userId)
	if err != nil {
		return nil, err
	}

	results := make(chan structs.OcrExport, len(fileDataResults))
	var wg sync.WaitGroup

	// Create a semaphore with a buffer size of 5
	semaphore := make(chan struct{}, 5)

	for _, fileData := range fileDataResults {
		wg.Add(1)
		go func(fd models.FileData) {
			defer wg.Done()

			// Acquire a semaphore slot
			semaphore <- struct{}{}

			filePath, err := fileRepository.BuildFilePath(simpleutils.UintToString(fd.ReceiptId), simpleutils.UintToString(fd.ID), fd.Name)
			if err != nil {
				results <- structs.OcrExport{OcrText: "", Filename: "", Err: err}
			} else {
				// TODO: V5: Refactor to use new ocr service
				systemReceiptProcessingSettings, err := repositories.SystemSettingsRepository{}.GetSystemReceiptProcessingSettings()
				if err != nil {
					results <- structs.OcrExport{OcrText: "", Filename: "", Err: err}
				}
				ocrService := NewOcrService(nil, systemReceiptProcessingSettings.ReceiptProcessingSettings)
				ocrText, err := ocrService.ReadImage(filePath)
				if err != nil {
					results <- structs.OcrExport{OcrText: "", Filename: "", Err: err}
				}
				results <- structs.OcrExport{OcrText: ocrText, Filename: fd.Name, Err: err}
			}

			// Release the semaphore slot
			<-semaphore
		}(fileData)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	ocrExportResults := make([]structs.OcrExport, 0)
	for r := range results {
		ocrExportResults = append(ocrExportResults, r)
	}

	return ocrExportResults, nil
}
