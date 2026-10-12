package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
)

// captureAuthSource runs a request through the middleware and reports the
// AuthSource the downstream handler would read off the claims.
func captureAuthSource(t *testing.T, request *http.Request) (int, string) {
	t.Helper()

	captured := ""

	recorder := httptest.NewRecorder()
	UnifiedAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = structs.GetClaims(r).AuthSource
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(recorder, request)

	return recorder.Result().StatusCode, captured
}

// TestUnifiedAuthMiddleware_MarksTheAuthSource pins the marker a handful of
// endpoints gate on.
//
// Some operations are ordinary for a signed-in human and an escalation for a
// key: connecting an identity provider above all, which turns a scoped,
// revocable credential into a permanent login. A handler can only tell the two
// apart if the middleware says which branch ran, and it is the only thing that
// knows.
func TestUnifiedAuthMiddleware_MarksTheAuthSource(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()
	t.Setenv("ENCRYPTION_KEY", "test-key")

	user := createTestUser()

	accessToken, _, _, err := services.GenerateJWT(user.ID)
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}

	bearer := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	bearer.Header.Set("Authorization", "Bearer "+accessToken)

	if status, source := captureAuthSource(t, bearer); status != http.StatusOK || source != structs.AuthSourceJwt {
		t.Errorf("a bearer JWT: expected 200 and %q, got %d and %q", structs.AuthSourceJwt, status, source)
	}

	// A plain navigation authenticates by cookie, and that is the shape a desktop
	// OIDC link arrives as -- so it has to be marked the same way the header is.
	cookie := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	cookie.AddCookie(&http.Cookie{Name: "jwt", Value: accessToken})

	if status, source := captureAuthSource(t, cookie); status != http.StatusOK || source != structs.AuthSourceJwt {
		t.Errorf("a JWT cookie: expected 200 and %q, got %d and %q", structs.AuthSourceJwt, status, source)
	}

	// Every scope, because the marker describes the CREDENTIAL and must not vary
	// with how much that credential is allowed to do.
	for _, scope := range []models.ApiKeyScope{
		models.API_KEY_SCOPE_READ,
		models.API_KEY_SCOPE_WRITE,
		models.API_KEY_SCOPE_READ_WRITE,
	} {
		_, rawKey, err := createTestApiKey(user.ID, string(scope))
		if err != nil {
			t.Fatalf("create api key: %v", err)
		}

		keyed := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		keyed.Header.Set("Authorization", rawKey)

		status, source := captureAuthSource(t, keyed)

		// A "w" key is refused this GET by the scope gate before any handler runs,
		// which is correct and not what this test is about -- so only assert the
		// marker where the request actually reaches a handler.
		if scope == models.API_KEY_SCOPE_WRITE {
			if status != http.StatusForbidden {
				t.Errorf("a %q key on a GET: expected 403 from the scope gate, got %d", scope, status)
			}
			continue
		}

		if status != http.StatusOK || source != structs.AuthSourceApiKey {
			t.Errorf("a %q key: expected 200 and %q, got %d and %q", scope, structs.AuthSourceApiKey, status, source)
		}
	}
}

// TestAuthSourceNeverRidesOnAToken is the reason the field is json:"-".
//
// Serialized, it would be settable by whoever minted the token and inheritable
// by any legacy one -- and "how did THIS request arrive" is not a property of a
// token at all. A signed JWT that claims to be an API key must still be read as
// what it is: an interactive session.
func TestAuthSourceNeverRidesOnAToken(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()

	planted := signClaims(t, structs.Claims{
		UserId:      user.ID,
		Username:    user.Username,
		Displayname: user.DisplayName,
		TokenType:   structs.TokenTypeAccess,
		AuthSource:  structs.AuthSourceApiKey,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	request.Header.Set("Authorization", "Bearer "+planted)

	status, source := captureAuthSource(t, request)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	if source != structs.AuthSourceJwt {
		t.Errorf("a claim planted in the token survived; expected %q, got %q", structs.AuthSourceJwt, source)
	}
}
