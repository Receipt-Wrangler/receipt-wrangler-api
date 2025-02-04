package wranglerasynq

import (
	"context"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"time"
)

func HandleRefreshTokenCleanupTask(context context.Context, task *asynq.Task) error {

	db := repositories.GetDB()
	return db.Model(&models.RefreshToken{}).
		Where("expires_at < ? OR is_used = ?", time.Now(), true).
		Delete(&models.RefreshToken{}).
		Error
}
