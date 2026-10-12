package wranglerasynq

import (
	"context"
	"testing"
	"time"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func TestHandleOidcCleanupTaskRemovesOnlySpentRows(t *testing.T) {
	defer repositories.TruncateTestDb()
	db := repositories.GetDB()

	live := models.OidcAuthSession{
		StateHash:    "live-state",
		NonceHash:    "live-nonce",
		CodeVerifier: "verifier",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	expired := models.OidcAuthSession{
		StateHash:    "expired-state",
		NonceHash:    "expired-nonce",
		CodeVerifier: "verifier",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	used := models.OidcAuthSession{
		StateHash:    "used-state",
		NonceHash:    "used-nonce",
		CodeVerifier: "verifier",
		Used:         true,
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	for _, session := range []models.OidcAuthSession{live, expired, used} {
		if err := db.Create(&session).Error; err != nil {
			t.Fatalf("failed to seed a session: %v", err)
		}
	}

	liveCode := models.OidcExchangeCode{CodeHash: "live-code", UserId: 1, CodeChallenge: "c", ExpiresAt: time.Now().Add(time.Hour)}
	expiredCode := models.OidcExchangeCode{CodeHash: "expired-code", UserId: 1, CodeChallenge: "c", ExpiresAt: time.Now().Add(-time.Hour)}

	for _, code := range []models.OidcExchangeCode{liveCode, expiredCode} {
		if err := db.Create(&code).Error; err != nil {
			t.Fatalf("failed to seed a code: %v", err)
		}
	}

	if err := HandleOidcCleanupTask(context.Background(), nil); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	var sessions []models.OidcAuthSession
	if err := db.Find(&sessions).Error; err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	if len(sessions) != 1 || sessions[0].StateHash != "live-state" {
		t.Errorf("expected only the live session to survive, got %+v", sessions)
	}

	var codes []models.OidcExchangeCode
	if err := db.Find(&codes).Error; err != nil {
		t.Fatalf("failed to list codes: %v", err)
	}

	if len(codes) != 1 || codes[0].CodeHash != "live-code" {
		t.Errorf("expected only the live code to survive, got %+v", codes)
	}
}
