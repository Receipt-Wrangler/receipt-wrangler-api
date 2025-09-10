package utils

import (
	"crypto/hmac"
	"crypto/sha256"
)

func GenerateHmac(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)

	return mac.Sum(nil)
}
