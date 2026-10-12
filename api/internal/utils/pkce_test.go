package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestVerifyPkceS256(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := challengeFor(verifier)

	tests := []struct {
		name      string
		verifier  string
		challenge string
		want      bool
	}{
		{"matching verifier and challenge", verifier, challenge, true},
		{"wrong verifier", "some-other-verifier", challenge, false},
		{"empty verifier", "", challenge, false},
		{"empty challenge", verifier, "", false},
		{"challenge is the plaintext verifier (plain not accepted)", verifier, verifier, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := VerifyPkceS256(test.verifier, test.challenge); got != test.want {
				t.Errorf("VerifyPkceS256(%q, %q) = %v, want %v", test.verifier, test.challenge, got, test.want)
			}
		})
	}
}
