package repositories

import (
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// TODO: Move to service
func (repository ReceiptImageRepository) CreateReceiptImage(fileData models.FileData, fileBytes []byte) (models.FileData, error) {
	fileRepository := NewFileRepository(repository.TX)
	db := repository.GetDB()

	// TODO: refactor to use command
	validatedFileType, err := fileRepository.ValidateFileType(fileBytes)
	if err != nil {
		return models.FileData{}, err
	}

	fileData.FileType = validatedFileType

	basePath, err := os.Getwd()
	if err != nil {
		return models.FileData{}, err
	}

	// Check if data path exists
	err = utils.DirectoryExists(basePath+"/data", true)
	if err != nil {
		return models.FileData{}, err
	}

	// Get initial group directory to see if it exists
	filePath, err := fileRepository.BuildFilePath(utils.UintToString(fileData.ReceiptId), "", fileData.Name)
	if err != nil {
		return models.FileData{}, err
	}
	groupDir, _ := filepath.Split(filePath)

	err = db.Model(models.FileData{}).Create(&fileData).Error
	if err != nil {
		os.Remove(filePath)
		return models.FileData{}, err
	}

	// Check if group's path exists
	err = utils.DirectoryExists(groupDir, true)
	if err != nil {
		return models.FileData{}, err
	}

	// Rebuild file path with correct file id
	filePath, err = fileRepository.BuildFilePath(utils.UintToString(fileData.ReceiptId), utils.UintToString(fileData.ID), fileData.Name)
	if err != nil {
		return models.FileData{}, err
	}

	err = utils.WriteFile(filePath, fileBytes)
	if err != nil {
		return models.FileData{}, err
	}

	return fileData, nil
}

func (repository ReceiptImageRepository) GetReceiptImageById(receiptImageId uint) (models.FileData, error) {
	db := repository.GetDB()
	var result models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Preload(clause.Associations).Find(&result).Error
	if err != nil {
		return models.FileData{}, err
	}

	return result, nil
}

func (repository ReceiptImageRepository) GetReceiptImagesByIdArray(receiptImageIds []uint) ([]models.FileData, error) {
	db := repository.GetDB()
	var result = make([]models.FileData, 0)

	err := db.Model(models.FileData{}).Where("id IN ?", receiptImageIds).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
