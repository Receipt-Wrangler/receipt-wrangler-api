package oidc

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

// tolerantBool accepts a JSON boolean and the string forms some identity
// providers emit instead.
//
// go-oidc solves this internally with an unexported stringAsBool, so a relying
// party decoding its own claims struct has to repeat it. Cognito is the usual
// culprit: it returns "email_verified": "true" as a string, and a plain bool
// field would fail the whole claim decode.
type tolerantBool bool

func (b *tolerantBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true", `"true"`, `"1"`, "1":
		*b = true
	case "false", `"false"`, `"0"`, "0", "null":
		*b = false
	default:
		return errors.New("invalid value for boolean")
	}

	return nil
}

// tolerantTime reads an epoch-seconds claim, as a JSON number (integer or
// fractional) or as a numeric string, the forms identity providers are seen to
// send. go-oidc exposes no auth_time field, so this relying party decodes it.
//
// It never fails the decode. Anything unreadable leaves it unset, which the one
// reader -- the delete confirmation's freshness check -- treats exactly like a
// provider that did not send auth_time at all. Failing instead would break every
// login for a provider that merely formats a claim oddly.
type tolerantTime struct {
	value *time.Time
}

func (t *tolerantTime) UnmarshalJSON(data []byte) error {
	raw := strings.Trim(strings.TrimSpace(string(data)), `"`)

	seconds, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		t.value = nil
		return nil
	}

	whole, fraction := math.Modf(seconds)
	parsed := time.Unix(int64(whole), int64(fraction*float64(time.Second)))
	t.value = &parsed

	return nil
}

// Time returns the claim's instant, or nil when it was absent or unreadable.
func (t *tolerantTime) Time() *time.Time {
	if t == nil {
		return nil
	}

	return t.value
}

// idTokenClaims are the claims this relying party reads.
//
// Only Subject is load-bearing for identity: it is the identity anchor and the
// only claim OIDC guarantees is stable and unique within an issuer. AuthTime is
// read by the delete confirmation alone, to check the provider really asked for a
// fresh sign-in. Everything else is a convenience for provisioning a display name
// or rendering a Connected Accounts row, and is refreshed on every login.
type idTokenClaims struct {
	Subject           string        `json:"sub"`
	PreferredUsername string        `json:"preferred_username"`
	Name              string        `json:"name"`
	Email             string        `json:"email"`
	EmailVerified     tolerantBool  `json:"email_verified"`
	AuthTime          *tolerantTime `json:"auth_time"`
}

func decodeIdTokenClaims(raw json.RawMessage) (idTokenClaims, error) {
	var claims idTokenClaims

	err := json.Unmarshal(raw, &claims)
	if err != nil {
		return idTokenClaims{}, err
	}

	claims.Subject = strings.TrimSpace(claims.Subject)
	claims.PreferredUsername = strings.TrimSpace(claims.PreferredUsername)
	claims.Name = strings.TrimSpace(claims.Name)
	claims.Email = strings.TrimSpace(claims.Email)

	return claims, nil
}

func splitScope(scope string) []string {
	return strings.Fields(scope)
}
