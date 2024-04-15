package services

import (
	"bytes"
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"gopkg.in/gographics/imagick.v2/imagick"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"
)

func ReadImage(path string) (string, error) {
	var text string
	appConfig := config.GetConfig()
	startTime := time.Now()

	imageBytes, err := prepareImage(path)
	if err != nil {
		return "", err
	}

	if appConfig.AiSettings.OcrEngine == structs.TESSERACT || appConfig.AiSettings.OcrEngine == "" {
		text, err = ReadImageWithTesseract(imageBytes)
		if err != nil {
			return "", err
		}
	}

	if appConfig.AiSettings.OcrEngine == structs.EASY_OCR {
		text, err = ReadImageWithEasyOcr(imageBytes)
		if err != nil {
			return "", err
		}
	}
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	logging.GetLogger().Print("OCR and Image processing took: ", elapsedTime)

	if appConfig.Debug.DebugOcr {
		err = writeDebuggingFiles(text, path, imageBytes, elapsedTime)
		if err != nil {
			return "", err
		}
	}

	return text, nil
}

func ReadImageWithTesseract(preparedImageBytes []byte) (string, error) {
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

func ReadImageWithEasyOcr(preparedImageBytes []byte) (string, error) {
	fileRepository := repositories.NewFileRepository(nil)
	tempPath, err := fileRepository.WriteTempFile(preparedImageBytes)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPath)

	var textBuffer bytes.Buffer
	var text string
	cmd := exec.Command("easyocr", "-l", "en", "-f", tempPath, "--detail", "0", "--gpu", "0")
	cmd.Stdout = &textBuffer

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	text = textBuffer.String()

	return text, nil
}

func writeDebuggingFiles(ocrText string, path string, imageBytes []byte, ocrDuration time.Duration) error {
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

func prepareImage(path string) ([]byte, error) {
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

	err = mw.ScaleImage(mw.GetImageWidth()/2, mw.GetImageHeight()/2)
	if err != nil {
		return nil, err
	}

	bytes := mw.GetImageBlob()

	return bytes, nil
}
