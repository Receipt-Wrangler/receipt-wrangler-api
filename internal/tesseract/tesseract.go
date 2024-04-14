package tesseract

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
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

func ReadImage(path string) (string, error) {
	var text string
	appConfig := config.GetConfig()

	imageBytes, err := prepareImage(path)
	if err != nil {
		return "", err
	}

	if appConfig.AiSettings.OcrEngine == structs.TESSERACT {
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

	/**
	TODO: Make configurable (choose your own OCR engine)
	TODO: Rename package
	TODO: Update sh file to get easy ocr
	TODO: update docs
	*/
	if appConfig.Debug.DebugOcr {
		err = writeDebuggingFiles(text, path, imageBytes)
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
	cmd := exec.Command("easyocr", "-l", "en", "-f", tempPath, "--detail", "0")
	cmd.Stdout = &textBuffer

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	text = textBuffer.String()

	return text, nil
}

func writeDebuggingFiles(ocrText string, path string, imageBytes []byte) error {
	fileRepository := repositories.NewFileRepository(nil)
	pathParts := strings.Split(path, "/")
	filename := pathParts[len(pathParts)-1]

	tempPath := fileRepository.GetTempDirectoryPath()
	textFilePath := filepath.Join(tempPath, filename+".txt")
	imageFilePath := filepath.Join(tempPath, filename+".jpg")

	textBytes := []byte(ocrText)

	err := utils.WriteFile(textFilePath, textBytes)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return err
	}

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

	err = mw.DeskewImage(0.10)
	if err != nil {
		return nil, err
	}

	bytes := mw.GetImageBlob()

	return bytes, nil
}
