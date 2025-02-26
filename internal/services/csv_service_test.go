package services

import (
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestShouldBuildCsv(t *testing.T) {
	expected :=
		"header1,header2\n" +
			"row1,row2\n" +
			"row3,row4\n"

	service := NewCsvService()
	headers := []string{"header1", "header2"}

	rows := [][]string{{"row1", "row2"}, {"row3", "row4"}}
	buffer, err := service.BuildCsv(headers, rows)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if buffer.String() != expected {
		utils.PrintTestError(t, "expected "+expected, buffer.String())
	}
}
