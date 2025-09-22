package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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

	aesBlock, err := aes.NewCipher(Sha256Hash128Bit(key))
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

	return Base64Encode(encryptedData), nil
}

func DecryptData(key string, encryptedData []byte) (string, error) {
	if len(key) == 0 {
		return "", errors.New("key cannot be empty")
	}
	if len(encryptedData) == 0 {
		return "", errors.New("encryptedData cannot be empty")
	}

	aesBlock, err := aes.NewCipher(Sha256Hash128Bit(key))
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
