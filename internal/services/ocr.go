package services

import (
	"bytes"
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"gopkg.in/gographics/imagick.v3/imagick"
	"gorm.io/gorm"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"
)

type OcrService struct {
	BaseService
	ReceiptProcessingSettings models.ReceiptProcessingSettings
}

func NewOcrService(tx *gorm.DB, receiptProcessingSettings models.ReceiptProcessingSettings) OcrService {
	service := OcrService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
		ReceiptProcessingSettings: receiptProcessingSettings,
	}

	return service
}

func (service OcrService) ReadImage(path string) (string, commands.UpsertSystemTaskCommand, error) {
	var text string
	startTime := time.Now()
	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:                 models.OCR_PROCESSING,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.RECEIPT_PROCESSING_SETTINGS,
		AssociatedEntityId:   service.ReceiptProcessingSettings.ID,
		StartedAt:            time.Now(),
	}

	imageBytes, err := service.prepareImage(path)
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
		return "", systemTaskCommand, err
	}

	if service.ReceiptProcessingSettings.OcrEngine != nil && *service.ReceiptProcessingSettings.OcrEngine == models.TESSERACT_NEW {
		text, err = service.ReadImageWithTesseract(imageBytes)
		if err != nil {
			systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
			systemTaskCommand.ResultDescription = err.Error()
			return "", systemTaskCommand, err
		}
	}

	if service.ReceiptProcessingSettings.OcrEngine != nil && *service.ReceiptProcessingSettings.OcrEngine == models.EASY_OCR_NEW {
		text, err = service.ReadImageWithEasyOcr(imageBytes)
		if err != nil {
			systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
			systemTaskCommand.ResultDescription = err.Error()
			return "", systemTaskCommand, err
		}
	}
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	logging.LogStd(logging.LOG_LEVEL_INFO, "OCR and Image processing took: ", elapsedTime)

	systemSettingsRepository := repositories.NewSystemSettingsRepository(service.TX)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		systemTaskCommand.Status = models.SYSTEM_TASK_FAILED
		systemTaskCommand.ResultDescription = err.Error()
		return "", systemTaskCommand, err
	}

	if systemSettings.DebugOcr {
		err = service.writeDebuggingFiles(text, path, imageBytes, elapsedTime)
		if err != nil {
			return "", commands.UpsertSystemTaskCommand{}, err
		}
	}

	ocrEndTime := time.Now()
	systemTaskCommand.EndedAt = &ocrEndTime
	systemTaskCommand.ResultDescription = text

	return text, systemTaskCommand, nil
}

func (service OcrService) ReadImageWithTesseract(preparedImageBytes []byte) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetVariable("tessedit_char_blacklist", "!@#$%^&*()_+=-[]}{;:'\"\\|~`<>/?")
	if err != nil {
		return "", nil
	}

	err = client.SetImageFromBytes(preparedImageBytes)
	if err != nil {
		return "", err
	}

	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

func (service OcrService) ReadImageWithEasyOcr(preparedImageBytes []byte) (string, error) {
	fileRepository := repositories.NewFileRepository(nil)
	tempPath, err := fileRepository.WriteTempFile(preparedImageBytes)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPath)

	var textBuffer bytes.Buffer
	var text string
	cmd := exec.Command("easyocr", "-l", "en", "-f", tempPath, "--detail", "0", "--gpu", "0", "--verbose", "0")
	cmd.Stdout = &textBuffer
	cmd.Stderr = io.Discard

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	text = textBuffer.String()

	return text, nil
}

func (service OcrService) writeDebuggingFiles(ocrText string, path string, imageBytes []byte, ocrDuration time.Duration) error {
	fileRepository := repositories.NewFileRepository(nil)
	pathParts := strings.Split(path, "/")
	filename := pathParts[len(pathParts)-1]

	tempPath := fileRepository.GetTempDirectoryPath()
	textFilePath := filepath.Join(tempPath, filename+".txt")
	imageFilePath := filepath.Join(tempPath, filename+".jpg")

	textBytes := []byte(ocrText)

	os.Remove(textFilePath)
	err := utils.WriteFile(textFilePath, textBytes)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return err
	}

	os.Remove(imageFilePath)
	imgFile, err := os.Create(imageFilePath)
	if err != nil {
		return err
	}
	defer imgFile.Close()

	err = jpeg.Encode(imgFile, img, nil)
	if err != nil {
		return err
	}

	err = utils.WriteFile(imageFilePath, imageBytes)
	if err != nil {
		return err
	}

	fmt.Println("OCR Text saved to: ", textFilePath)
	fmt.Println("OCR Image saved to: ", imageFilePath)
	fmt.Println("OCR and image processing duration: ", ocrDuration)

	return nil
}

func (service OcrService) prepareImage(path string) ([]byte, error) {
	mw := imagick.NewMagickWand()
	err := mw.ReadImage(path)
	if err != nil {
		return nil, err
	}

	err = mw.TrimImage(0)
	if err != nil {
		return nil, err
	}

	err = mw.SetImageType(imagick.IMAGE_TYPE_BILEVEL)
	if err != nil {
		return nil, err
	}

	err = mw.BlurImage(0, 1.5)
	if err != nil {
		return nil, err
	}

	err = mw.SharpenImage(0, 1)
	if err != nil {
		return nil, err
	}

	err = mw.EnhanceImage()
	if err != nil {
		return nil, err
	}

	err = mw.ContrastImage(false)
	if err != nil {
		return nil, err
	}

	err = mw.DeskewImage(.40)
	if err != nil {
		return nil, err
	}

	if service.ReceiptProcessingSettings.OcrEngine != nil &&
		*service.ReceiptProcessingSettings.OcrEngine == models.EASY_OCR_NEW {
		err = mw.ScaleImage(mw.GetImageWidth()/2, mw.GetImageHeight()/2)
		if err != nil {
			return nil, err
		}
	}

	return mw.GetImageBlob()
}
