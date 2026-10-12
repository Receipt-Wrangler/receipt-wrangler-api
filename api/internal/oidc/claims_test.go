package oidc

import (
	"encoding/json"
	"testing"
)

// TestTolerantBoolAcceptsTheStringForms covers the gap go-oidc leaves: it solves
// this internally with an unexported stringAsBool, so a relying party decoding
// its own claims struct has to repeat it. Cognito returns email_verified as a
// string, and a plain bool field would fail the entire claim decode.
func TestTolerantBoolAcceptsTheStringForms(t *testing.T) {
	tests := []struct {
		raw      string
		expected bool
		wantErr  bool
	}{
		{`true`, true, false},
		{`false`, false, false},
		{`"true"`, true, false},
		{`"false"`, false, false},
		{`1`, true, false},
		{`0`, false, false},
		{`"1"`, true, false},
		{`"0"`, false, false},
		{`null`, false, false},
		{`"yes"`, false, true},
		{`{}`, false, true},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			var value tolerantBool
			err := json.Unmarshal([]byte(test.raw), &value)

			if test.wantErr {
				if err == nil {
					t.Errorf("expected %s to be rejected", test.raw)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected %s to decode, got %v", test.raw, err)
			}

			if bool(value) != test.expected {
				t.Errorf("expected %v, got %v", test.expected, bool(value))
			}
		})
	}
}

func TestDecodeIdTokenClaimsTrimsAndSurvivesAStringEmailVerified(t *testing.T) {
	raw := json.RawMessage(`{
		"sub": "  abc123  ",
		"preferred_username": " hank ",
		"name": " Hank Hill ",
		"email": " hank@example.com ",
		"email_verified": "true"
	}`)

	claims, err := decodeIdTokenClaims(raw)
	if err != nil {
		t.Fatalf("expected the claims to decode, got %v", err)
	}

	if claims.Subject != "abc123" {
		t.Errorf("expected the subject to be trimmed, got %q", claims.Subject)
	}

	if claims.PreferredUsername != "hank" {
		t.Errorf("expected preferred_username to be trimmed, got %q", claims.PreferredUsername)
	}

	if !bool(claims.EmailVerified) {
		t.Error("expected a string email_verified to decode as true")
	}
}

func TestDecodeIdTokenClaimsToleratesMissingOptionalClaims(t *testing.T) {
	// Only sub is load-bearing; a provider that sends nothing else must still work.
	claims, err := decodeIdTokenClaims(json.RawMessage(`{"sub":"only-a-subject"}`))
	if err != nil {
		t.Fatalf("expected the claims to decode, got %v", err)
	}

	if claims.Subject != "only-a-subject" {
		t.Errorf("unexpected subject %q", claims.Subject)
	}
}
