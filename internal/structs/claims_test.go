package structs

import (
	"context"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaims_Validate_Success(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestClaims_Validate_AllValidFields(t *testing.T) {
	claims := Claims{
		UserId:             123,
		Username:           "fulluser",
		Displayname:        "Full Display Name",
		UserRole:           models.USER,
		DefaultAvatarColor: "#00FF00",
		ApiKeyScope:        models.API_KEY_SCOPE_READ_WRITE,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
}

func TestClaims_Validate_MissingUserId(t *testing.T) {
	claims := Claims{
		UserId:             0, // Invalid - should be > 0
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for missing user ID")
	}

	expectedMsg := "user ID is required"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_MissingUsername(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "", // Invalid - should not be empty
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for missing username")
	}

	expectedMsg := "username is required"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_MissingDisplayname(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "", // Invalid - should not be empty
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for missing display name")
	}

	expectedMsg := "display name is required"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_InvalidUserRole(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.UserRole("INVALID_ROLE"), // Invalid role
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid user role")
	}

	expectedMsg := "invalid user role: INVALID_ROLE"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_ValidUserRoles(t *testing.T) {
	// Test ADMIN role
	claimsAdmin := Claims{
		UserId:             1,
		Username:           "admin",
		Displayname:        "Admin User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claimsAdmin.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error for ADMIN role")
	}

	// Test USER role
	claimsUser := Claims{
		UserId:             2,
		Username:           "user",
		Displayname:        "Regular User",
		UserRole:           models.USER,
		DefaultAvatarColor: "#00FF00",
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err = claimsUser.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error for USER role")
	}
}

func TestClaims_Validate_InvalidAvatarColorFormat(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "FF0000", // Invalid - missing #
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid avatar color format")
	}

	expectedMsg := "invalid avatar color format: FF0000"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_InvalidAvatarColorLength(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF00", // Invalid - too short
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid avatar color length")
	}

	expectedMsg := "invalid avatar color format: #FF00"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_InvalidHexColor(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#GGGGGG", // Invalid - G is not a hex digit
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid hex color")
	}

	expectedMsg := "invalid hex color: #GGGGGG"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_EmptyAvatarColor(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "", // Empty should be allowed
		ApiKeyScope:        models.API_KEY_SCOPE_READ,
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error for empty avatar color")
	}
}

func TestClaims_Validate_ValidAvatarColors(t *testing.T) {
	validColors := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFFFF", "#000000", "#123ABC"}

	for _, color := range validColors {
		claims := Claims{
			UserId:             1,
			Username:           "testuser",
			Displayname:        "Test User",
			UserRole:           models.ADMIN,
			DefaultAvatarColor: color,
			RegisteredClaims:   jwt.RegisteredClaims{},
		}

		err := claims.Validate(context.Background())
		if err != nil {
			utils.PrintTestError(t, err, "no error for valid color: "+color)
		}
	}
}

func TestClaims_Validate_InvalidApiKeyScope(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        models.ApiKeyScope("invalid_scope"), // Invalid scope
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err == nil {
		utils.PrintTestError(t, err, "an error for invalid API key scope")
	}

	expectedMsg := "invalid API key scope: invalid_scope"
	if err.Error() != expectedMsg {
		utils.PrintTestError(t, err.Error(), expectedMsg)
	}
}

func TestClaims_Validate_ValidApiKeyScopes(t *testing.T) {
	validScopes := []models.ApiKeyScope{models.API_KEY_SCOPE_READ, models.API_KEY_SCOPE_WRITE, models.API_KEY_SCOPE_READ_WRITE}

	for _, scope := range validScopes {
		claims := Claims{
			UserId:             1,
			Username:           "testuser",
			Displayname:        "Test User",
			UserRole:           models.ADMIN,
			DefaultAvatarColor: "#FF0000",
			ApiKeyScope:        scope,
			RegisteredClaims:   jwt.RegisteredClaims{},
		}

		err := claims.Validate(context.Background())
		if err != nil {
			utils.PrintTestError(t, err, "no error for valid scope: "+string(scope))
		}
	}
}

func TestClaims_Validate_EmptyApiKeyScope(t *testing.T) {
	claims := Claims{
		UserId:             1,
		Username:           "testuser",
		Displayname:        "Test User",
		UserRole:           models.ADMIN,
		DefaultAvatarColor: "#FF0000",
		ApiKeyScope:        "", // Empty should be allowed
		RegisteredClaims:   jwt.RegisteredClaims{},
	}

	err := claims.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error for empty API key scope")
	}
}

func TestClaims_Validate_AllMinimalFields(t *testing.T) {
	// Test with minimal required fields only
	claims := Claims{
		UserId:           1,
		Username:         "minimaluser",
		Displayname:      "Minimal User",
		UserRole:         models.USER,
		RegisteredClaims: jwt.RegisteredClaims{},
		// DefaultAvatarColor and ApiKeyScope are optional and empty
	}

	err := claims.Validate(context.Background())
	if err != nil {
		utils.PrintTestError(t, err, "no error for minimal valid claims")
	}
}
