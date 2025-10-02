package models

import (
	"database/sql/driver"
	"errors"
)

type OcrEngine string

const (
	TESSERACT     OcrEngine = "tesseract"
	TESSERACT_NEW OcrEngine = "TESSERACT"
)

func (ocrEngine *OcrEngine) Scan(value string) error {
	*ocrEngine = OcrEngine(value)
	return nil
}

func (ocrEngine OcrEngine) Value() (driver.Value, error) {
	if len(ocrEngine) == 0 {
		return "", nil
	}

	if ocrEngine != TESSERACT && ocrEngine != TESSERACT_NEW {
		return nil, errors.New("invalid ocr type")
	}
	return string(ocrEngine), nil
}
