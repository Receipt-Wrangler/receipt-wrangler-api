package services

import (
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/tesseract"
	"receipt-wrangler/api/internal/utils"
)

func ReadReceiptImage(receiptImageId string) (models.Receipt, error) {
	var result models.Receipt
	var pathToReadFrom string

	receiptImageUint, err := simpleutils.StringToUint(receiptImageId)
	if err != nil {
		return result, err
	}

	receiptImageRepository := repositories.NewReceiptImageRepository(nil)
	receiptImage, err := receiptImageRepository.GetReceiptImageById(receiptImageUint)
	if err != nil {
		return result, err
	}
	fileRepository := repositories.NewReceiptImageRepository(nil)

	receiptImagePath, err := fileRepository.BuildFilePath(simpleutils.UintToString(receiptImage.ReceiptId), receiptImageId, receiptImage.Name)
	if err != nil {
		return result, err
	}

	receiptImageBytes, err := utils.ReadFile(receiptImagePath)
	if err != nil {
		return models.Receipt{}, err
	}

	// TODO: make generic
	if receiptImage.FileType == constants.APPLICATION_PDF {
		bytes, err := fileRepository.ConvertPdfToJpg(receiptImageBytes)
		if err != nil {
			return models.Receipt{}, err
		}

		pathToReadFrom, err = fileRepository.WriteTempFile(bytes)
		if err != nil {
			return models.Receipt{}, err
		}

		defer os.Remove(pathToReadFrom)
	} else {
		pathToReadFrom = receiptImagePath
	}

	ocrText, err := tesseract.ReadImage(pathToReadFrom)
	if err != nil {
		return result, nil
	}

	result, err = ReadReceiptData(ocrText)
	if err != nil {
		return models.Receipt{}, err
	}

	return result, nil
}

func ReadReceiptImageFromFileOnly(path string) (models.Receipt, error) {
	var result models.Receipt

	ocrText, err := tesseract.ReadImage(path)
	if err != nil {
		return result, nil
	}

	result, err = ReadReceiptData(ocrText)
	if err != nil {
		return models.Receipt{}, err
	}

	return result, nil
}

func MagicFillFromImage(command commands.MagicFillCommand) (models.Receipt, error) {
	fileRepository := repositories.NewFileRepository(nil)

	bytes, err := fileRepository.GetBytesFromImageBytes(command.ImageData)
	if err != nil {
		return models.Receipt{}, err
	}

	filePath, err := fileRepository.WriteTempFile(bytes)
	if err != nil {
		return models.Receipt{}, err
	}

	filledReceipt, err := ReadReceiptImageFromFileOnly(filePath)
	if err != nil {
		os.Remove(filePath)
		return models.Receipt{}, nil
	}

	os.Remove(filePath)
	return filledReceipt, nil
}
