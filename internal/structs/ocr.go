package structs

import (
	"database/sql/driver"
	"errors"
)

type OcrExport struct {
	OcrText  string
	Filename string
	Err      error
}

type OcrEngine string

const (
	TESSERACT OcrEngine = "tesseract"
	EASY_OCR  OcrEngine = "easyOcr"
)

func (ocrEngine *OcrEngine) Scan(value string) error {
	*ocrEngine = OcrEngine(value)
	return nil
}

func (ocrEngine OcrEngine) Value() (driver.Value, error) {
	if ocrEngine != TESSERACT || ocrEngine != EASY_OCR {
		return nil, errors.New("invalid ocr type")
	}
	return string(ocrEngine), nil
}
