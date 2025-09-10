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

func Sha256Hash128Bit(valueToHash string) []byte {
	hashedValue := Sha256Hash([]byte(valueToHash))
	return []byte(hashedValue[:16])
}
