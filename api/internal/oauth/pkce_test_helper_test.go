package oauth

import (
	"crypto/sha256"
	"encoding/base64"
)

// challengeFor derives the S256 code challenge for a verifier, independently of
// utils.VerifyPkceS256 so the OAuth flow tests are not judged by the same code
// they exercise.
func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
