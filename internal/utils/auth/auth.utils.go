package auth_utils

import (
	"context"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers/auth"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func InitJwtValidator() (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		config := config.GetConfig()
		return []byte(config.SecretKey), nil
	}

	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS512,
		"https://recieptWrangler.io",
		[]string{"https://receiptWrangler.io"},
		// validator.WithCustomClaims(), TODO: Add for custom claim validation
		validator.WithAllowedClockSkew(30*time.Second),
	)

	return jwtValidator, err
}

func InitTokenRefreshValidator() (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		config := config.GetConfig()
		return []byte(config.SecretKey), nil
	}

	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS512,
		"https://recieptWrangler.io",
		[]string{"https://receiptWrangler.io"},
		validator.WithCustomClaims(customClaims),
		validator.WithAllowedClockSkew(24*time.Hour),
	)

	return jwtValidator, err
}

func customClaims() validator.CustomClaims {
	return &auth.Claims{}
}

// func customClaims() validator.CustomClaims {
// 	return &Claims{}
// }
