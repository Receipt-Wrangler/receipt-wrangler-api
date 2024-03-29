package tesseract

import (
	"bytes"
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"gopkg.in/gographics/imagick.v2/imagick"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

func ReadImage(path string) (string, error) {
	config := config.GetConfig()
	client := gosseract.NewClient()
	defer client.Close()

	imageBytes, err := prepareImage(path)
	if err != nil {
		return "", err
	}

	err = client.SetImageFromBytes(imageBytes)
	if err != nil {
		return "", err
	}

	text, err := client.Text()
	if err != nil {
		return "", err
	}

	if config.Debug.DebugOcr {
		err = writeDebuggingFiles(text, path, imageBytes)
		if err != nil {
			return "", err
		}
	}

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

	bytes := mw.GetImageBlob()

	return bytes, nil
}
