package repositories

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"gopkg.in/gographics/imagick.v3/imagick"
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

func (repository FileRepository) BuildFilePath(receiptId string, receiptImageId string, receiptImageFileName string) (string, error) {
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

	fileName := utils.BuildFileName(receiptId, receiptImageId, receiptImageFileName)
	path := filepath.Join(groupPath, fileName)

	return path, nil
}

func (repository FileRepository) BuildGroupPath(groupId uint, alternateGroupName string) (string, error) {
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

	strGroupId := utils.UintToString(groupId)
	groupPath, err := utils.BuildGroupPathString(strGroupId, groupNameToUse)
	if err != nil {
		return "", err
	}

	return groupPath, nil
}

func (repository FileRepository) GetBytesForFileData(fileData models.FileData) ([]byte, error) {
	path, err := repository.BuildFilePath(utils.UintToString(fileData.ReceiptId), utils.UintToString(fileData.ID), fileData.Name)

	if err != nil {
		return nil, err
	}

	fileBytes, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}

	resultBytes, err := repository.GetBytesFromImageBytes(fileBytes)
	if err != nil {
		return nil, err
	}

	return resultBytes, nil
}

func (repository FileRepository) GetBytesFromImageBytes(imageData []byte) ([]byte, error) {
	var bytes []byte
	validatedType, err := repository.ValidateFileType(imageData)
	if err != nil {
		return nil, err
	}

	if validatedType == constants.ApplicationPdf {
		bytes, err = repository.ConvertPdfToJpg(imageData)
		if err != nil {
			return nil, err
		}
	} else if validatedType == constants.ImageHeic {
		bytes, err = repository.ConvertHeicToJpg(imageData)
		if err != nil {
			return nil, err
		}
	} else {
		bytes = imageData
	}

	return bytes, nil
}

func (repository FileRepository) IsImage(imageData []byte) (bool, error) {
	validatedFileType, err := repository.ValidateFileType(imageData)
	if err != nil {
		return false, err
	}

	isImage, err := regexp.Match(constants.AnyImage, []byte(validatedFileType))
	if err != nil {
		return false, err
	}

	return isImage, nil
}

func (repository FileRepository) IsPdf(imageData []byte) (bool, error) {
	validatedFileType, err := repository.ValidateFileType(imageData)
	if err != nil {
		return false, err
	}

	isPdf, err := regexp.Match(constants.ApplicationPdf, []byte(validatedFileType))
	if err != nil {
		return false, err
	}

	return isPdf, nil
}

func (repository FileRepository) ValidateFileType(bytes []byte) (string, error) {
	fileType := utils.GetMimeType(bytes).String()
	acceptedFileTypes := []string{constants.AnyImage, constants.ApplicationPdf}

	for _, acceptedFileType := range acceptedFileTypes {
		matched, _ := regexp.Match(acceptedFileType, []byte(fileType))

		if matched {
			return fileType, nil
		}
	}

	return "", errors.New("invalid file type")
}

func (repository FileRepository) ValidateJsonFileType(bytes []byte) (string, error) {
	fileType := mimetype.Detect(bytes).String()
	acceptedFileTypes := []string{constants.ApplicationJson}

	for _, acceptedFileType := range acceptedFileTypes {
		matched, _ := regexp.Match(acceptedFileType, []byte(fileType))

		if matched {
			return fileType, nil
		}
	}

	return "", errors.New("invalid file type")
}

func (repository FileRepository) ConvertHeicToJpg(bytes []byte) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(bytes); err != nil {
		return nil, err
	}

	if err := mw.SetImageFormat("jpeg"); err != nil {
		return nil, err
	}

	return mw.GetImageBlob()
}

func (repository FileRepository) ConvertPdfToJpg(bytes []byte) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(bytes); err != nil {
		return nil, err
	}

	// Set the format to JPEG once, the setting is retained across frames.
	if err := mw.SetImageFormat("jpeg"); err != nil {
		return nil, err
	}

	// Must be *after* ReadImageFile
	// Flatten image and remove alpha channel, to prevent alpha turning black in jpg
	if err := mw.SetImageAlphaChannel(0); err != nil {
		return nil, err
	}

	// Find out how many images/pages we've got in a pdf.
	numPages := int(mw.GetNumberImages())

	// Create a new wand to store the final long image.
	finalImage := imagick.NewMagickWand()
	defer finalImage.Destroy()

	// Iterate over each page, processing it as needed.
	for i := 0; i < numPages; i++ {
		mw.SetIteratorIndex(i)

		// Get the current image as a MagickWand.
		// This is done because AddImage() expects a MagickWand, not a blob.
		currImage := mw.GetImage()

		// Add the current image to the finalImage wand.
		if err := finalImage.AddImage(currImage); err != nil {
			log.Fatal(err) // Handle the error appropriately
		}

		// Destroy the current image object as it's no longer needed.
		currImage.Destroy()
	}

	// Now, we will append all the images stored in finalImage vertically.
	// Resetting the wand is necessary for AppendImages to work.
	finalImage.ResetIterator()
	combinedImage := finalImage.AppendImages(true)

	tempFilePath, err := repository.BuildTempFilePath("jpg")
	if err != nil {
		return nil, err
	}

	if err := combinedImage.WriteImage(tempFilePath); err != nil {
		return nil, err
	}

	bytes, err = utils.ReadFile(tempFilePath)
	if err != nil {
		return nil, err
	}

	os.Remove(tempFilePath)
	return bytes, nil
}

func (repository FileRepository) WriteTempFile(data []byte) (string, error) {
	tempPath := repository.GetTempDirectoryPath()
	utils.MakeDirectory(tempPath)

	validatedFileType, err := repository.ValidateFileType(data)
	if err != nil {
		return "", err
	}

	parts := strings.Split(validatedFileType, "/")
	if len(parts) != 2 {
		return "", errors.New("malformed mime type")
	}

	fileType := parts[1]

	filePath, err := repository.BuildTempFilePath(fileType)
	if err != nil {
		return "", err
	}

	err = utils.WriteFile(filePath, data)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	return filePath, nil
}

func (repository FileRepository) BuildTempFilePath(fileType string) (string, error) {
	tempPath := repository.GetTempDirectoryPath()

	filename, err := utils.GetRandomString(10)
	if err != nil {
		return "", err
	}

	filePath := tempPath + "/" + filename
	filePath = filePath + "." + fileType
	return filePath, nil
}

func (repository FileRepository) GetFileType(bytes []byte) (string, error) {
	fileType, err := repository.ValidateFileType(bytes)
	if err != nil {
		return "", err
	}

	isPdf, err := repository.IsPdf(bytes)
	if err != nil {
		return "", err
	}

	if isPdf {
		fileType = "image/jpeg"
	}

	return fileType, nil
}

func (repository FileRepository) BuildEncodedImageString(bytes []byte) (string, error) {
	fileType, err := repository.GetFileType(bytes)
	if err != nil {
		return "", err
	}

	return utils.BuildDataURI(fileType, bytes), nil
}

func (repository FileRepository) CreateZipFromTempFiles(zipFilename string, filenames []string) (string, error) {
	tempPath := repository.GetTempDirectoryPath()
	zipPath := filepath.Join(tempPath, zipFilename)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filename := range filenames {
		filePath := filepath.Join(tempPath, filename)
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return "", err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return "", err
		}

		header.Name = filename
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return "", err
		}

		if _, err = io.Copy(writer, file); err != nil {
			return "", err
		}
	}

	return zipPath, nil
}

func (repository FileRepository) ZipFiles(filenames []string, fileContents [][]byte) ([]byte, error) {
	if len(filenames) != len(fileContents) {
		return nil, errors.New("number of filenames does not match number of file contents")
	}

	if len(filenames) == 0 {
		return nil, errors.New("no files to zip")
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	for i, filename := range filenames {
		file, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("error creating file %s in zip: %w", filename, err)
		}

		_, err = file.Write(fileContents[i])
		if err != nil {
			return nil, fmt.Errorf("error writing content for file %s in zip: %w", filename, err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing zip writer: %w", err)
	}

	return zipBuffer.Bytes(), nil
}

func (repository FileRepository) GetTempDirectoryPath() string {
	return config.GetBasePath() + "/temp"
}

func (repository FileRepository) GetTestJpgBytes() ([]byte, error) {
	path := config.GetBasePath() + "/testing/test.jpg"

	return utils.ReadFile(path)
}
