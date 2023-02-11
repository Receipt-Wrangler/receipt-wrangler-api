package structs

import (
	"receipt-wrangler/api/internal/models"
	"time"
)

type APIUser struct {
	ID          uint            `json:"id"`
	DisplayName string          `json:"displayName"`
	Username    string          `json:"username"`
	UserRole    models.UserRole `json:"userRole"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}
