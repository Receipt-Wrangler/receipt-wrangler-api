package utils

import (
	"context"
	"receipt-wrangler/api/internal/models"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserId      uint
	Username    string
	Displayname string
	UserRole models.UserRole
	jwt.RegisteredClaims
}

func (claim *Claims) Validate(ctx context.Context) error { // TODO: Implement claim validation
	return nil
}
