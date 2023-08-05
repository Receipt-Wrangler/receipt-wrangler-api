package services

import (
	"receipt-wrangler/api/internal/ai"
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

	repository := repositories.NewReceiptImageRepository(nil)
	receiptImage, err := repository.GetReceiptImageById(receiptImageUint)
	if err != nil {
		return result, err
	}

	path, err := utils.BuildFilePath(simpleutils.UintToString(receiptImage.ReceiptId), receiptImageId, receiptImage.Name)
	if err != nil {
		return result, err
	}

	ocrText, err := tesseract.ReadImage(path)
	if err != nil {
		return result, nil
	}

	result, err = ai.ReadReceiptData(ocrText)
	if err != nil {
		return models.Receipt{}, err
	}


	return result, nil
}
