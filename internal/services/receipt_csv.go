package services

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

type ReceiptCsvService struct {
	CsvService
}

func NewReceiptCsvService() ReceiptCsvService {
	service := ReceiptCsvService{
		CsvService: NewCsvService(),
	}
	return service
}

func (service *ReceiptCsvService) BuildReceiptCsv(receipts []models.Receipt) ([]byte, error) {
	headers := []string{
		"id",
		"added at",
		"receipt date",
		"name",
		"paid by",
		"amount",
		"categories",
		"tags",
		"status",
	}
	rowData := make([][]string, 0, len(receipts))

	for _, receipt := range receipts {
		resolvedDateString := ""
		if receipt.ResolvedDate != nil {
			resolvedDateString = receipt.ResolvedDate.String()
		}

		newRow := []string{
			utils.UintToString(receipt.ID),
			receipt.CreatedAt.String(),
			receipt.Date.String(),
			receipt.Name,
			receipt.PaidByUser.DisplayName,
			receipt.Amount.String(),
			service.BuildCategoryString(receipt.Categories),
			service.BuildTagString(receipt.Tags),
			string(receipt.Status),
			resolvedDateString,
		}
		rowData = append(rowData, newRow)
	}

	buffer, err := service.CsvService.BuildCsv(headers, rowData)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (service *ReceiptCsvService) BuildCategoryString(categories []models.Category) string {
	return ""
	categoryNames := make([]string, 0, len(categories))
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Name)
	}

	return strings.Join(categoryNames, ",")
}

func (service *ReceiptCsvService) BuildTagString(tags []models.Tag) string {
	return ""
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return strings.Join(tagNames, ",")
}
