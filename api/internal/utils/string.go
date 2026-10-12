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

	return Base64URLEncode(bytes), nil
}

// GetRandomUrlSafeString returns byteLength bytes of cryptographic randomness
// encoded as unpadded base64url, safe to embed verbatim in a URL query
// parameter. Unpadded encoding avoids the '=' characters that GetRandomString
// (base64.URLEncoding) emits and that would otherwise need escaping.
//
// Used for OAuth client ids and authorization codes (internal/oauth) and for
// OIDC state, nonce, browser-binding and exchange-code values (internal/oidc).
func GetRandomUrlSafeString(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func RemoveJsonFormat(input string) string {
	result := input
	result = strings.ReplaceAll(result, "```json", "")
	result = strings.ReplaceAll(result, "```", "")

	return result
}
