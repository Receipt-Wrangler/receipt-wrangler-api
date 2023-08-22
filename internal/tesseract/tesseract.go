package tesseract

import (
	"github.com/otiai10/gosseract/v2"
	"gopkg.in/gographics/imagick.v2/imagick"
)

var client *gosseract.Client

func InitClient() {
	client = gosseract.NewClient()
}

func GetClient() *gosseract.Client {
	return client
}

func ReadImage(path string) (string, error) {
	bytes, err := prepareImage(path)
	if err != nil {
		return "", err
	}

	err = client.SetImageFromBytes(bytes)
	if err != nil {
		return "", err
	}

	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
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
