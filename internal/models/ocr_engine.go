package models

import (
	"database/sql/driver"
	"errors"
)

type OcrEngine string

const (
	TESSERACT     OcrEngine = "tesseract"
	EASY_OCR      OcrEngine = "easyOcr"
	TESSERACT_NEW OcrEngine = "TESSERACT"
	EASY_OCR_NEW  OcrEngine = "EASY_OCR"
)

func (ocrEngine *OcrEngine) Scan(value string) error {
	*ocrEngine = OcrEngine(value)
	return nil
}

func (ocrEngine OcrEngine) Value() (driver.Value, error) {
	if len(ocrEngine) == 0 {
		return "", nil
	}

	if ocrEngine != TESSERACT && ocrEngine != EASY_OCR && ocrEngine != TESSERACT_NEW && ocrEngine != EASY_OCR_NEW {
		return nil, errors.New("invalid ocr type")
	}
	return string(ocrEngine), nil
}
