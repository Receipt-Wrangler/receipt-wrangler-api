package structs

import (
	"context"
	"fmt"
	"receipt-wrangler/api/internal/models"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	DefaultAvatarColor string             `json:"defaultAvatarColor"`
	Displayname        string             `json:"displayName"`
	UserId             uint               `json:"userId"`
	Username           string             `json:"username"`
	UserRole           models.UserRole    `json:"userRole"`
	ApiKeyScope        models.ApiKeyScope `json:"apiKeyScope"`
	jwt.RegisteredClaims
}

func (claim *Claims) Validate(ctx context.Context) error {
	if claim.UserId == 0 {
		return fmt.Errorf("user ID is required")
	}

	if claim.Username == "" {
		return fmt.Errorf("username is required")
	}

	if claim.Displayname == "" {
		return fmt.Errorf("display name is required")
	}

	// Validate UserRole is a valid enum value
	validRoles := []models.UserRole{models.ADMIN, models.USER}
	roleValid := false
	for _, role := range validRoles {
		if claim.UserRole == role {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return fmt.Errorf("invalid user role: %s", claim.UserRole)
	}

	// Validate DefaultAvatarColor format (should be hex color)
	if claim.DefaultAvatarColor != "" {
		if !strings.HasPrefix(claim.DefaultAvatarColor, "#") || len(claim.DefaultAvatarColor) != 7 {
			return fmt.Errorf("invalid avatar color format: %s", claim.DefaultAvatarColor)
		}
		// Check if it's valid hex
		if _, err := strconv.ParseInt(claim.DefaultAvatarColor[1:], 16, 64); err != nil {
			return fmt.Errorf("invalid hex color: %s", claim.DefaultAvatarColor)
		}
	}

	// Validate API key scope if present (should be valid scope values)
	if claim.ApiKeyScope != "" {
		if !claim.ApiKeyScope.IsValid() {
			return fmt.Errorf("invalid API key scope: %s", claim.ApiKeyScope)
		}
	}

	return nil
}
