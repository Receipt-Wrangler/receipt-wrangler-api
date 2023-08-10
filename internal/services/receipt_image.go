package services

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/tesseract"
)

func ReadReceiptImage(receiptImageId string) (models.Receipt, error) {
	var result models.Receipt
	receiptImageUint, err := simpleutils.StringToUint(receiptImageId)
	if err != nil {
		return result, err
	}

	repository := repositories.NewReceiptImageRepository(nil)
	receiptImage, err := repository.GetReceiptImageById(receiptImageUint)
	if err != nil {
		return result, err
	}

	path, err := repositories.BuildFilePath(simpleutils.UintToString(receiptImage.ReceiptId), receiptImageId, receiptImage.Name)
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
