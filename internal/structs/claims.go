package structs

import (
	"context"
	"receipt-wrangler/api/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	DefaultAvatarColor string          `json:"defaultAvatarColor"`
	Displayname        string          `json:"displayName"`
	UserId             uint            `json:"userId"`
	Username           string          `json:"username"`
	UserRole           models.UserRole `json:"userRole"`
	ApiKeyScope        string          `json:"apiKeyScope"`
	jwt.RegisteredClaims
}

func (claim *Claims) Validate(ctx context.Context) error { // TODO: Implement claim validation
	return nil
}
