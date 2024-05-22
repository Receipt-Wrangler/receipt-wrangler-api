package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func EncryptData(key string, value []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("encryption key cannot be empty, please set the environment variable: ENCRYPTION_KEY")
	}
	if len(value) == 0 {
		return nil, errors.New("value cannot be empty")
	}

	aesBlock, err := aes.NewCipher([]byte(Md5Hash(key)))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, value, nil), nil
}

func EncryptAndEncodeToBase64(key string, value string) (string, error) {
	encryptedData, err := EncryptData(key, []byte(value))
	if err != nil {
		return "", err
	}

	return EncodeToBase64(encryptedData), nil
}

func DecryptData(key string, encryptedData []byte) (string, error) {
	if len(key) == 0 {
		return "", errors.New("key cannot be empty")
	}
	if len(encryptedData) == 0 {
		return "", errors.New("encryptedData cannot be empty")
	}

	aesBlock, err := aes.NewCipher([]byte(Md5Hash(key)))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	nonce, cipherText := encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():]
	clearText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(clearText), nil
}

func DecryptB64EncodedData(key string, encodedCipherText string) (string, error) {
	decodedCipherTextBytes, err := Base64Decode(encodedCipherText)
	if err != nil {
		return "", err
	}

	cleartext, err := DecryptData(key, decodedCipherTextBytes)
	if err != nil {
		return "", err
	}

	return cleartext, nil
}

func Md5Hash(valueToHash string) string {
	bytesValue := []byte(valueToHash)
	hashedValue := md5.Sum(bytesValue)
	return string(hashedValue[:])
}

func EncodeToBase64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func Base64Decode(b64EncodedValue string) ([]byte, error) {
	result, err := base64.StdEncoding.DecodeString(b64EncodedValue)
	if err != nil {
		return nil, err
	}

	return result, nil
}
