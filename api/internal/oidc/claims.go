package oidc

import (
	"encoding/json"
	"errors"
	"strings"
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

// idTokenClaims are the claims this relying party reads.
//
// Only Subject is load-bearing: it is the identity anchor and the only claim
// OIDC guarantees is stable and unique within an issuer. Everything else is a
// convenience for provisioning a display name or rendering a Connected Accounts
// row, and is refreshed on every login.
type idTokenClaims struct {
	Subject           string       `json:"sub"`
	PreferredUsername string       `json:"preferred_username"`
	Name              string       `json:"name"`
	Email             string       `json:"email"`
	EmailVerified     tolerantBool `json:"email_verified"`
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
