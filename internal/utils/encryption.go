package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func EncryptData(key string, value []byte) ([]byte, error) {
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

func DecryptData(key string, encryptedData []byte) (string, error) {
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

func Md5Hash(valueToHash string) string {
	bytesValue := []byte(valueToHash)
	hashedValue := md5.Sum(bytesValue)
	return string(hashedValue[:])
}

func EncodeToBase64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}
