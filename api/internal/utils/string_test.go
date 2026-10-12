package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGetRandomStringShouldReturnDecodableStringOfRequestedByteLength(t *testing.T) {
	length := 32

	random, err := GetRandomString(length)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	decoded, err := Base64URLDecode(random)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	if len(decoded) != length {
		PrintTestError(t, len(decoded), length)
	}
}

func TestGetRandomStringShouldReturnUniqueValues(t *testing.T) {
	first, err := GetRandomString(16)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	second, err := GetRandomString(16)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	if first == second {
		t.Errorf("Expected two random strings to differ, got identical values")
	}
}

func TestRemoveJsonFormatShouldStripFences(t *testing.T) {
	input := "```json\n{\"a\":1}\n```"
	expected := "\n{\"a\":1}\n"

	actual := RemoveJsonFormat(input)
	if actual != expected {
		PrintTestError(t, actual, expected)
	}

	if strings.Contains(actual, "```") {
		t.Errorf("Expected all code fences to be removed, got %q", actual)
	}
}

func TestRemoveJsonFormatShouldLeavePlainStringUnchanged(t *testing.T) {
	input := "just a plain string"

	actual := RemoveJsonFormat(input)
	if actual != input {
		PrintTestError(t, actual, input)
	}
}

// TestGetRandomUrlSafeStringShouldReturnUnpaddedBase64UrlOfRequestedByteLength
// pins the two properties the callers depend on: the full requested entropy
// survives the encoding, and the result is unpadded.
//
// Decoded with base64.RawURLEncoding, NOT utils.Base64URLDecode -- and that is
// the point rather than an inconvenience. Base64URLDecode is base64.URLEncoding,
// the PADDED decoder, so the two are interchangeable only when byteLength is a
// multiple of 3. They are not interchangeable at 32, which is exactly
// oidc.randomTokenBytes: the length behind every state, nonce, browser binding,
// exchange code and link-launch handle. Writing this test against
// Base64URLDecode would therefore either fail outright or, worse, quietly start
// passing if someone switched the function to padded encoding -- which would
// break embedding these values in a URL, the whole reason the function exists.
func TestGetRandomUrlSafeStringShouldReturnUnpaddedBase64UrlOfRequestedByteLength(t *testing.T) {
	// The lengths actually requested in the codebase: OAuth client ids (24),
	// every OIDC token (32) and a provisioned account's discarded password (48).
	// 32 is the one of the three that produces a padding difference, so dropping
	// it would let a padded implementation pass.
	for _, length := range []int{24, 32, 48} {
		random, err := GetRandomUrlSafeString(length)
		if err != nil {
			PrintTestError(t, err, nil)
			continue
		}

		decoded, err := base64.RawURLEncoding.DecodeString(random)
		if err != nil {
			PrintTestError(t, err, nil)
			continue
		}

		if len(decoded) != length {
			PrintTestError(t, len(decoded), length)
		}

		// Unpadded base64url: no '=' to escape, and none of standard base64's
		// '+' or '/', which would change meaning in a query string.
		if strings.ContainsAny(random, "=+/") {
			PrintTestError(t, random, "a string free of '=', '+' and '/'")
		}
	}
}

func TestGetRandomUrlSafeStringShouldReturnUniqueValues(t *testing.T) {
	first, err := GetRandomUrlSafeString(32)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	second, err := GetRandomUrlSafeString(32)
	if err != nil {
		PrintTestError(t, err, nil)
	}

	if first == second {
		PrintTestError(t, second, "a value different from "+first)
	}
}
