package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256Hash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
