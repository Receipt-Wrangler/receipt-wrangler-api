package wranglerasynq

import (
	"context"
	"receipt-wrangler/api/internal/repositories"

	"github.com/hibiken/asynq"
)

// HandleOidcCleanupTask removes spent OIDC auth sessions, mobile exchange codes
// and mobile link launch handles.
//
// This is hygiene rather than a security control: all three tables are already
// guarded by expires_at inside the consume statements' WHERE clauses, so an
// un-swept row can never be redeemed. Rows are consumed within seconds or expire
// within minutes, so the tables stay small at any interval.
func HandleOidcCleanupTask(context context.Context, task *asynq.Task) error {
	return repositories.NewOidcSessionRepository(nil).DeleteExpiredOidcSessions()
}
