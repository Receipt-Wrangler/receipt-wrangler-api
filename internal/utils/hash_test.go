package utils

import "testing"

func TestSha256Hash128Bit(t *testing.T) {
	value := "superSecretData"
	expected := "4M2yAEADbol2mSGOXAMLNA=="

	hashedValue := Sha256Hash128Bit(value)
	hashedString := Base64Encode(hashedValue)

	if hashedString != expected {
		PrintTestError(t, hashedString, expected)
	}

	if len(hashedValue) != 16 {
		PrintTestError(t, len(hashedValue), 16)
	}
}
