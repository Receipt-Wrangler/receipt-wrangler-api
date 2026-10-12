package oidc

import (
	"sync"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

// TestConsumeAuthSessionIsAtomic is the guard on the callback's very first step.
//
// The used and expiry checks live in the UPDATE's WHERE clause so check-and-set
// is one statement. A read-then-update would let two concurrent callbacks bearing
// the same state both observe it unused and both proceed to a token exchange.
func TestConsumeAuthSessionIsAtomic(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	created, err := createAuthSession(newAuthSessionParams{
		ProviderId:   provider.ID,
		ClientType:   models.OidcClientDesktop,
		CodeVerifier: "a-verifier",
	})
	if err != nil {
		t.Fatalf("failed to create the session: %v", err)
	}

	var waitGroup sync.WaitGroup
	results := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			_, claimed, err := consumeAuthSession(created.State)
			if err != nil {
				results <- false
				return
			}

			results <- claimed
		}()
	}

	waitGroup.Wait()
	close(results)

	claimedCount := 0
	for claimed := range results {
		if claimed {
			claimedCount++
		}
	}

	if claimedCount != 1 {
		t.Errorf("exactly one caller must claim the session, got %d", claimedCount)
	}
}

func TestConsumeExchangeCodeIsAtomic(t *testing.T) {
	defer teardownOidcTest()
	repositories.CreateTestRoles()

	user := createTestUser(t, "atomicexchange")

	code, err := createExchangeCode(user.ID, challengeFor("a-verifier"))
	if err != nil {
		t.Fatalf("failed to create the code: %v", err)
	}

	var waitGroup sync.WaitGroup
	results := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			consumed, err := consumeExchangeCode(code)
			if err != nil {
				results <- false
				return
			}

			results <- consumed
		}()
	}

	waitGroup.Wait()
	close(results)

	consumedCount := 0
	for consumed := range results {
		if consumed {
			consumedCount++
		}
	}

	if consumedCount != 1 {
		t.Errorf("exactly one caller must consume the code, got %d", consumedCount)
	}
}

func TestDeleteExpiredOidcSessionsKeepsLiveRows(t *testing.T) {
	defer teardownOidcTest()
	_, provider := setupOidcTest(t, oidcTestOptions{})

	live, err := createAuthSession(newAuthSessionParams{
		ProviderId:   provider.ID,
		ClientType:   models.OidcClientDesktop,
		CodeVerifier: "live-verifier",
	})
	if err != nil {
		t.Fatalf("failed to create the live session: %v", err)
	}

	expired, err := createAuthSession(newAuthSessionParams{
		ProviderId:   provider.ID,
		ClientType:   models.OidcClientDesktop,
		CodeVerifier: "expired-verifier",
	})
	if err != nil {
		t.Fatalf("failed to create the expiring session: %v", err)
	}

	if err := repositories.GetDB().
		Model(&models.OidcAuthSession{}).
		Where("state_hash = ?", hashSecret(expired.State)).
		Update("expires_at", timeInPast()).Error; err != nil {
		t.Fatalf("failed to expire the session: %v", err)
	}

	if err := repositories.NewOidcSessionRepository(nil).DeleteExpiredOidcSessions(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	var remaining int64
	if err := repositories.GetDB().Model(&models.OidcAuthSession{}).Count(&remaining).Error; err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}

	if remaining != 1 {
		t.Fatalf("expected only the live session to remain, found %d", remaining)
	}

	if _, claimed, err := consumeAuthSession(live.State); err != nil || !claimed {
		t.Errorf("the live session should still be claimable (claimed=%v, err=%v)", claimed, err)
	}
}
