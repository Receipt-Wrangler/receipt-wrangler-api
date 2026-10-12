package oidc

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func runExchange(t *testing.T, code string, verifier string) (*httptest.ResponseRecorder, error) {
	t.Helper()

	body := `{"code":"` + code + `","codeVerifier":"` + verifier + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/oidc/exchange", bytes.NewBufferString(body))
	recorder := httptest.NewRecorder()

	_, err := Exchange(recorder, request)

	return recorder, err
}

// seedExchangeCode issues a code bound to the given verifier's challenge.
func seedExchangeCode(t *testing.T, userId uint, verifier string) string {
	t.Helper()

	code, err := createExchangeCode(userId, challengeFor(verifier))
	if err != nil {
		t.Fatalf("failed to create the exchange code: %v", err)
	}

	return code
}

func TestExchangeReturnsTokensInTheBodyAndNeverACookie(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "exchangeuser")
	verifier := "an-app-generated-verifier-long-enough-for-pkce"
	code := seedExchangeCode(t, user.ID, verifier)

	recorder, err := runExchange(t, code, verifier)
	if err != nil {
		t.Fatalf("expected the exchange to succeed, got %v", err)
	}

	// This mirrors login?tokensInBody=true exactly, so the mobile client reuses its
	// existing storeAppData path unchanged.
	if len(recorder.Result().Cookies()) > 0 {
		t.Error("the exchange endpoint must never set a cookie")
	}
}

func TestExchangeRejectsAWrongVerifier(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "wrongverifier")
	code := seedExchangeCode(t, user.ID, "the-real-verifier-held-only-by-the-app")

	_, err := runExchange(t, code, "a-verifier-an-interceptor-guessed")
	if !errors.Is(err, ErrInvalidExchange) {
		t.Fatalf("expected ErrInvalidExchange, got %v", err)
	}
}

// TestExchangeDoesNotConsumeTheCodeOnAWrongVerifier is why the code is loaded and
// verified BEFORE it is burned.
//
// Burning it first would let anyone who intercepted the redirect destroy a valid
// code out from under the real app just by presenting a wrong verifier -- a free
// denial of service on every sign-in attempt.
func TestExchangeDoesNotConsumeTheCodeOnAWrongVerifier(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "notburned")
	verifier := "the-real-verifier-held-only-by-the-app"
	code := seedExchangeCode(t, user.ID, verifier)

	if _, err := runExchange(t, code, "wrong"); !errors.Is(err, ErrInvalidExchange) {
		t.Fatalf("expected the wrong verifier to be refused, got %v", err)
	}

	var stored models.OidcExchangeCode
	if err := repositories.GetDB().Where("code_hash = ?", hashSecret(code)).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload the code: %v", err)
	}

	if stored.Used {
		t.Fatal("a failed proof must not burn the code")
	}

	// And the real app can still redeem it.
	if _, err := runExchange(t, code, verifier); err != nil {
		t.Errorf("the genuine app should still be able to redeem the code, got %v", err)
	}
}

func TestExchangeCodeIsSingleUse(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "singleuse")
	verifier := "an-app-generated-verifier-long-enough-for-pkce"
	code := seedExchangeCode(t, user.ID, verifier)

	if _, err := runExchange(t, code, verifier); err != nil {
		t.Fatalf("the first redemption should succeed, got %v", err)
	}

	if _, err := runExchange(t, code, verifier); !errors.Is(err, ErrInvalidExchange) {
		t.Errorf("expected the second redemption to be refused, got %v", err)
	}
}

func TestExchangeRejectsAnExpiredCode(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "expiredcode")
	verifier := "an-app-generated-verifier-long-enough-for-pkce"
	code := seedExchangeCode(t, user.ID, verifier)

	if err := repositories.GetDB().
		Model(&models.OidcExchangeCode{}).
		Where("code_hash = ?", hashSecret(code)).
		Update("expires_at", timeInPast()).Error; err != nil {
		t.Fatalf("failed to expire the code: %v", err)
	}

	if _, err := runExchange(t, code, verifier); !errors.Is(err, ErrInvalidExchange) {
		t.Errorf("expected an expired code to be refused, got %v", err)
	}
}

func TestExchangeRejectsAnUnknownCode(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	// Unknown, expired, used and wrong-verifier are deliberately indistinguishable,
	// so a caller cannot probe which codes exist.
	if _, err := runExchange(t, "a-code-that-was-never-issued", "whatever"); !errors.Is(err, ErrInvalidExchange) {
		t.Errorf("expected ErrInvalidExchange, got %v", err)
	}
}
