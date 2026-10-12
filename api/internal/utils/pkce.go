package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

// VerifyPkceS256 reports whether codeVerifier matches the stored S256
// codeChallenge per RFC 7636: BASE64URL-WITHOUT-PADDING(SHA256(verifier)).
// Only the S256 method is supported; "plain" is intentionally rejected by
// callers before reaching here. The comparison is constant time.
//
// Shared by the OAuth 2.1 authorization server (internal/oauth, verifying a
// client's verifier at the token endpoint) and the OIDC relying party
// (internal/oidc, verifying the mobile app's verifier at the exchange
// endpoint). It lives here so the two cannot drift.
func VerifyPkceS256(codeVerifier string, codeChallenge string) bool {
	if len(codeVerifier) == 0 || len(codeChallenge) == 0 {
		return false
	}

	sum := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])

	return SecureCompare(computed, codeChallenge)
}

// SecureCompare reports whether two strings are equal in constant time. Use it
// for any secret comparison -- a byte-by-byte == leaks how much of a guess was
// right through its timing.
func SecureCompare(a string, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
