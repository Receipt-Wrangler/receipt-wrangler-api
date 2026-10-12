package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// fakeIdp is a real OpenID Connect provider served over httptest: discovery,
// a JWKS, and a token endpoint that mints genuinely RS256-signed ID tokens.
//
// It exists so the callback tests exercise go-oidc for real -- signature
// verification, JWKS fetching, issuer and audience checks -- rather than stubbing
// the one library whose behavior this feature's security depends on. It uses
// golang-jwt/jwt/v5, already a direct dependency, so it adds nothing to go.mod.
type fakeIdp struct {
	server    *httptest.Server
	key       *rsa.PrivateKey
	keyId     string
	clientId  string
	discovery int

	mu sync.Mutex
	// claims is what the next minted ID token will carry. Tests mutate this to
	// exercise a wrong nonce, a missing subject, and so on.
	claims jwt.MapClaims
	// omitIdToken makes the token endpoint return an OAuth-only response, which a
	// relying party must refuse.
	omitIdToken bool
	// signWithWrongKey mints a token signed by a key not in the JWKS.
	signWithWrongKey bool
}

func newFakeIdp(t *testing.T, clientId string) *fakeIdp {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	idp := &fakeIdp{key: key, keyId: "test-key", clientId: clientId}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", idp.handleDiscovery)
	mux.HandleFunc("/jwks", idp.handleJwks)
	mux.HandleFunc("/token", idp.handleToken)
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {})

	idp.server = httptest.NewServer(mux)
	t.Cleanup(idp.server.Close)

	return idp
}

func (idp *fakeIdp) issuer() string {
	return idp.server.URL
}

// discoveryCount reports how many times the discovery document was fetched, which
// is what the provider-cache tests assert on.
func (idp *fakeIdp) discoveryCount() int {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	return idp.discovery
}

func (idp *fakeIdp) setClaims(claims jwt.MapClaims) {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	idp.claims = claims
}

func (idp *fakeIdp) setOmitIdToken(omit bool) {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	idp.omitIdToken = omit
}

func (idp *fakeIdp) setSignWithWrongKey(wrong bool) {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	idp.signWithWrongKey = wrong
}

// baseClaims is a well-formed ID token payload for this issuer and client.
func (idp *fakeIdp) baseClaims(subject string, nonce string) jwt.MapClaims {
	return jwt.MapClaims{
		"iss":                idp.issuer(),
		"aud":                idp.clientId,
		"sub":                subject,
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
		"iat":                time.Now().Unix(),
		"nonce":              nonce,
		"preferred_username": "",
		"name":               "",
		"email":              "",
	}
}

func (idp *fakeIdp) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	idp.mu.Lock()
	idp.discovery++
	idp.mu.Unlock()

	writeJson(w, map[string]any{
		"issuer":                                idp.issuer(),
		"authorization_endpoint":                idp.issuer() + "/auth",
		"token_endpoint":                        idp.issuer() + "/token",
		"jwks_uri":                              idp.issuer() + "/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (idp *fakeIdp) handleJwks(w http.ResponseWriter, r *http.Request) {
	pub := idp.key.Public().(*rsa.PublicKey)

	writeJson(w, map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"alg": "RS256",
			"use": "sig",
			"kid": idp.keyId,
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}},
	})
}

func (idp *fakeIdp) handleToken(w http.ResponseWriter, r *http.Request) {
	idp.mu.Lock()
	claims := idp.claims
	omit := idp.omitIdToken
	wrongKey := idp.signWithWrongKey
	idp.mu.Unlock()

	response := map[string]any{
		"access_token": "fake-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
	}

	if !omit {
		signingKey := idp.key
		if wrongKey {
			other, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			signingKey = other
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = idp.keyId

		signed, err := token.SignedString(signingKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response["id_token"] = signed
	}

	writeJson(w, response)
}

func writeJson(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}
