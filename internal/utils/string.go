package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func GetRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func RemoveJsonFormat(input string) string {
	result := input
	result = strings.ReplaceAll(result, "```json", "")
	result = strings.ReplaceAll(result, "```", "")

	return result
}
