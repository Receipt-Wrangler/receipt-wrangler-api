package repositories

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"
	"regexp"

	"gopkg.in/gographics/imagick.v2/imagick"
	"gorm.io/gorm"
)

type FileRepository struct {
	BaseRepository
}

func NewFileRepository(tx *gorm.DB) FileRepository {
	repository := FileRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository BaseRepository) BuildFilePath(receiptId string, receiptImageId string, receiptImageFileName string) (string, error) {
	db := repository.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).Select("group_id").Find(&receipt).Error
	if err != nil {
		return "", err
	}

	groupPath, err := repository.BuildGroupPath(receipt.GroupId, "")
	if err != nil {
		return "", err
	}

	fileName := simpleutils.BuildFileName(receiptId, receiptImageId, receiptImageFileName)
	path := filepath.Join(groupPath, fileName)

	return path, nil
}

func (repository BaseRepository) BuildGroupPath(groupId uint, alternateGroupName string) (string, error) {
	db := repository.GetDB()
	var groupNameToUse string

	if len(alternateGroupName) > 0 {
		groupNameToUse = alternateGroupName
	} else {
		var group models.Group
		err := db.Model(models.Group{}).Where("id = ?", groupId).Select("name").Find(&group).Error
		if err != nil {
			return "", err
		}

		groupNameToUse = group.Name
	}

	strGroupId := simpleutils.UintToString(groupId)
	groupPath, err := simpleutils.BuildGroupPathString(strGroupId, groupNameToUse)
	if err != nil {
		return "", err
	}

	return groupPath, nil
}

func (repository BaseRepository) GetBytesForFileData(fileData models.FileData) ([]byte, error) {
	path, err := repository.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), simpleutils.UintToString(fileData.ID), fileData.Name)
	var bytes []byte

	if err != nil {
		return nil, err
	}

	isImage, err := repository.IsImage(fileData)
	if err != nil {
		return nil, err
	}

	isPdf, err := repository.IsPdf(fileData)
	if err != nil {
		return nil, err
	}

	if isPdf {
		bytes, err = repository.ConvertPdfToJpg(path)
		if err != nil {
			return nil, err
		}

	} else if isImage {
		bytes, err = utils.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("invalid file type")
	}

	return bytes, nil
}

func (repository BaseRepository) IsImage(fileData models.FileData) (bool, error) {
	isImage, err := regexp.Match(constants.ANY_IMAGE, []byte(fileData.FileType))
	if err != nil {
		return false, err
	}

	return isImage, nil
}

func (repository BaseRepository) IsPdf(fileData models.FileData) (bool, error) {
	isPdf, err := regexp.Match(constants.APPLICATION_PDF, []byte(fileData.FileType))
	if err != nil {
		return false, err
	}

	return isPdf, nil
}

func (repository BaseRepository) ValidateFileType(fileData models.FileData) (string, error) {
	fileType := http.DetectContentType(fileData.ImageData)
	acceptedFileTypes := []string{constants.ANY_IMAGE, constants.APPLICATION_PDF}

	for _, acceptedFileType := range acceptedFileTypes {
		matched, _ := regexp.Match(acceptedFileType, []byte(fileType))

		if matched {
			return fileType, nil
		}
	}

	return "", errors.New("invalid file type")
}

func (repository BaseRepository) ConvertPdfToJpg(filePath string) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Must be *before* ReadImageFile
	// Make sure our image is high quality
	if err := mw.SetResolution(300, 300); err != nil {
		return nil, err
	}

	// Load the image file into imagick
	if err := mw.ReadImage(filePath); err != nil {
		return nil, err
	}

	// Must be *after* ReadImageFile
	// Flatten image and remove alpha channel, to prevent alpha turning black in jpg
	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN); err != nil {
		return nil, err
	}

	// Set any compression (100 = max quality)
	if err := mw.SetCompressionQuality(95); err != nil {
		return nil, err
	}

	// Select only first page of pdf
	mw.SetIteratorIndex(0)

	// Convert into JPG
	if err := mw.SetFormat("jpg"); err != nil {
		return nil, err
	}

	// Save File

	mw.ResetIterator()
	return mw.GetImageBlob(), nil
}

func (repository BaseRepository) WriteTempFile(filename string, data []byte) (string, error) {
	tempPath := config.GetBasePath() + "/temp"
	utils.MakeDirectory(tempPath)

	filePath := tempPath + "/" + filename

	err := utils.WriteFile(filePath, data)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	return filePath, nil
}
