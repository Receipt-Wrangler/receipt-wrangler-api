package services

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"
	"time"
)

type ReceiptService struct {
	BaseService
}

func NewReceiptService(tx *gorm.DB) ReceiptService {
	service := ReceiptService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service ReceiptService) GetReceiptByReceiptImageId(receiptImageId string) (models.Receipt, error) {
	db := service.GetDB()
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Select("receipt_id").First(&fileData).Error
	if err != nil {
		return models.Receipt{}, err
	}

	receiptRepository := repositories.NewReceiptRepository(service.TX)
	receipt, err := receiptRepository.GetReceiptById(strconv.FormatUint(uint64(fileData.ReceiptId), 10))
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (service ReceiptService) DeleteReceipt(id string) error {
	db := service.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload("ImageFiles").Find(&receipt).Error
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var imagesToDelete []string
		fileRepository := repositories.NewFileRepository(tx)
		fileRepository.SetTransaction(tx)

		for _, f := range receipt.ImageFiles {
			path, _ := fileRepository.BuildFilePath(utils.UintToString(f.ReceiptId), utils.UintToString(f.ID), f.Name)
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

func (service ReceiptService) QuickScan(
	token *structs.Claims,
	paidByUserId uint,
	groupId uint,
	status models.ReceiptStatus,
	tempPath string,
	originalFileName string,
	asynqTaskId string,
) (models.Receipt, error) {
	db := repositories.GetDB()
	systemTaskService := NewSystemTaskService(service.TX)
	systemTaskRepository := repositories.NewSystemTaskRepository(service.TX)
	var createdReceipt models.Receipt

	fileRepository := repositories.NewFileRepository(service.TX)
	fileBytes, err := utils.ReadFile(tempPath)
	if err != nil {
		return models.Receipt{}, err
	}

	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		return models.Receipt{}, err
	}

	validatedFileType, err := fileRepository.ValidateFileType(fileBytes)
	if err != nil {
		return models.Receipt{}, err
	}

	magicFillCommand := commands.MagicFillCommand{
		ImageData: fileBytes,
		Filename:  originalFileName,
	}

	receiptRepository := repositories.NewReceiptRepository(service.TX)
	receiptImageRepository := repositories.NewReceiptImageRepository(service.TX)

	groupIdString := utils.UintToString(groupId)

	now := time.Now()
	receiptCommand, receiptProcessingMetadata, err := MagicFillFromImage(magicFillCommand, groupIdString)
	finishedAt := time.Now()

	metaCombineSystemTask, err := systemTaskRepository.CreateSystemTask(commands.UpsertSystemTaskCommand{
		Type:                 models.META_COMBINE_QUICK_SCAN,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.NOOP_ENTITY_TYPE,
		AssociatedEntityId:   0,
		AsynqTaskId:          asynqTaskId,
	})
	if err != nil {
		return models.Receipt{}, err
	}

	quickScanSystemTasks, taskErr := systemTaskService.CreateSystemTasksFromMetadata(
		receiptProcessingMetadata,
		now,
		finishedAt,
		models.QUICK_SCAN,
		&token.UserId,
		func(command commands.UpsertSystemTaskCommand) *uint {
			return &metaCombineSystemTask.ID
		})
	if taskErr != nil {
		return models.Receipt{}, taskErr
	}

	if err != nil {
		return models.Receipt{}, err
	}

	if receiptCommand.PaidByUserID == 0 {
		receiptCommand.PaidByUserID = paidByUserId
	}

	if len(receiptCommand.Status) == 0 {
		receiptCommand.Status = models.ReceiptStatus(status)
	}

	receiptCommand.GroupId = groupId

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository.SetTransaction(tx)
		receiptImageRepository.SetTransaction(tx)
		systemTaskService.SetTransaction(tx)
		uploadStart := time.Now()

		createdReceipt, err = receiptRepository.CreateReceipt(receiptCommand, token.UserId)
		uploadEnd := time.Now()
		_, taskErr := systemTaskService.CreateReceiptUploadedSystemTask(
			err,
			createdReceipt,
			quickScanSystemTasks,
			uploadStart,
		)
		if taskErr != nil {
			return taskErr
		}
		if err != nil {
			tx.Commit()
			return err
		}

		fileData := models.FileData{
			Name:      originalFileName,
			Size:      uint(fileInfo.Size()),
			ReceiptId: createdReceipt.ID,
			FileType:  validatedFileType,
		}
		_, err := receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
		if err != nil {
			return err
		}

		err = systemTaskService.AssociateSystemTasksToReceipt(
			createdReceipt.ID,
			metaCombineSystemTask.ID,
			uploadStart,
			uploadEnd)
		if err != nil {
			tx.Commit()
			return err
		}

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	os.Remove(tempPath)
	return createdReceipt, nil
}
