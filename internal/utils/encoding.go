package utils

import (
	"encoding/base64"
)

// Base64Encode encodes bytes to standard base64
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes standard base64 string
func Base64Decode(encoded string) ([]byte, error) {
	result, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Base64URLEncode encodes bytes to URL-safe base64
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes URL-safe base64 string
func Base64URLDecode(encoded string) ([]byte, error) {
	result, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Base64EncodeString is a convenience function for encoding strings
func Base64EncodeString(s string) string {
	return Base64Encode([]byte(s))
}

// BuildDataURI creates a data URI for the given MIME type and data
func BuildDataURI(mimeType string, data []byte) string {
	return "data:" + mimeType + ";base64," + Base64Encode(data)
}
