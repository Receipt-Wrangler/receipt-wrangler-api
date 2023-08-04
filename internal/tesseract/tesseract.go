package tesseract

import "github.com/otiai10/gosseract/v2"

var client *gosseract.Client

func InitClient() {
	client = gosseract.NewClient()
}

func GetClient() *gosseract.Client {
	return client
}

func ReadImage(path string) (string, error) {
	client.SetImage(path)
	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}
