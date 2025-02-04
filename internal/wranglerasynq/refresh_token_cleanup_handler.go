package wranglerasynq

import (
	"context"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"time"
)

func HandleRefreshTokenCleanupTask(context context.Context, task *asynq.Task) error {
	// NOTE: Length check can be removed in a while. Length check was added to remove refresh tokens that were not hashed
	db := repositories.GetDB()
	return db.Model(&models.RefreshToken{}).
		Where("expires_at < ? OR is_used = ? OR length(token) > 64", time.Now(), true).
		Delete(&models.RefreshToken{}).
		Error
}
