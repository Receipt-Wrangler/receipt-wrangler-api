package services

import (
	"os"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/tesseract"
	"receipt-wrangler/api/internal/utils"
)

func ReadReceiptImage(receiptImageId string) (models.Receipt, error) {
	var result models.Receipt
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
	path, err := fileRepository.BuildFilePath(simpleutils.UintToString(receiptImage.ReceiptId), receiptImageId, receiptImage.Name)
	if err != nil {
		return result, err
	}

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
	var magicFillCommand commands.MagicFillCommand

	tempPath := config.GetBasePath() + "/temp"
	utils.MakeDirectory(tempPath)

	filePath := tempPath + "/" + magicFillCommand.Filename
	utils.WriteFile(filePath, []byte(magicFillCommand.ImageData))

	filledReceipt, err := ReadReceiptImageFromFileOnly(filePath)
	if err != nil {
		return models.Receipt{}, nil
	}

	os.Remove(filePath)

	return filledReceipt, nil
}
