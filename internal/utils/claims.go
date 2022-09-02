package utils

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Username    string
	Displayname string
	jwt.RegisteredClaims
}

func (claim *Claims) Validate(ctx context.Context) error { // TODO: Implement claim validation
	return nil
}
