package utils

import (
	"testing"
)

func TestShouldEncryptStringWithAES128(t *testing.T) {
	key := "superSecureKey"
	value := []byte("superSecretData")

	cipherText, err := EncryptData(key, value)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	encodedCipherText := EncodeToBase64(cipherText)

	if len(encodedCipherText) != 60 {
		PrintTestError(t, len(cipherText), 60)
	}
}

func TestShouldDecryptStringWithAES128(t *testing.T) {
	key := "superSecureKey"
	value := []byte("superSecretData")

	cipherText, err := EncryptData(key, value)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	encodedCipherText := EncodeToBase64(cipherText)

	if len(encodedCipherText) != 60 {
		PrintTestError(t, len(cipherText), 60)
	}

	clearText, err := DecryptData(key, []byte(cipherText))
	if err != nil {
		PrintTestError(t, err, nil)
	}

	if clearText != "superSecretData" {
		PrintTestError(t, clearText, "superSecretData")
	}
}
