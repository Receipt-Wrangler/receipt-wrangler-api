package utils

import (
	"context"
	"fmt"
	"os"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	os.Args = append(os.Args, "-env=test")
	SetUpTestEnv()

	defer func() {
	}()

	return m.Run(), nil
}

func TestInitTokenValidatorReturnsValidator(t *testing.T) {
	v, err := InitTokenValidator()

	if v == nil {
		PrintTestError(t, v, "instance of validator")
	}

	if err != nil {
		PrintTestError(t, err, nil)
	}
}

func TestGenerateJWTGeneratesJWTCorrectly(t *testing.T) {
	expectedDisplayname := "Displayname"
	expectedUsername := "Test"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		PrintTestError(t, err, nil)
	}

	db := db.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		PrintTestError(t, err.Error(), nil)
	}

	jwt, _, _, err := GenerateJWT(user.ID)
	if err != nil {
		PrintTestError(t, jwt, "jwt token")
	}

	rawJwtStruct, err := v.ValidateToken(context.Background(), jwt)
	if err != nil {
		PrintTestError(t, rawJwtStruct, "claim object")
	}

	jwtClaims := rawJwtStruct.(*validator.ValidatedClaims).CustomClaims.(*Claims)

	if jwt == "nil" {
		PrintTestError(t, jwt, "non empty string")
	}

	if jwtClaims.UserId != user.ID {
		PrintTestError(t, jwtClaims.UserId, user.ID)
	}

	if jwtClaims.Displayname != expectedDisplayname {
		PrintTestError(t, jwtClaims.Displayname, expectedDisplayname)
	}

	if jwtClaims.Username != expectedUsername {
		PrintTestError(t, jwtClaims.Username, expectedUsername)
	}

	if jwtClaims.Issuer != expectedIssuer {
		PrintTestError(t, jwtClaims.Issuer, expectedIssuer)
	}

	if len(jwtClaims.Audience) > 0 && jwtClaims.Audience[0] != expectedIssuer {
		PrintTestError(t, jwtClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		PrintTestError(t, err, nil)
	}
}

func TestGenerateRefreshTokenCorrectly(t *testing.T) {
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		PrintTestError(t, err, nil)
	}

	db := db.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		PrintTestError(t, err.Error(), nil)
	}

	_, refreshToken, _, err := GenerateJWT(user.ID)
	if err != nil {
		PrintTestError(t, refreshToken, "refresh token")
	}

	rawRefreshTokenClaims, err := v.ValidateToken(context.Background(), refreshToken)
	if err != nil {
		PrintTestError(t, rawRefreshTokenClaims, "claim object")
	}

	refreshTokenClaims := rawRefreshTokenClaims.(*validator.ValidatedClaims).CustomClaims.(*Claims)

	if refreshToken == "nil" {
		PrintTestError(t, refreshToken, "non empty string")
	}

	if refreshTokenClaims.UserId != user.ID {
		PrintTestError(t, refreshTokenClaims.UserId, user.ID)
	}

	if refreshTokenClaims.Issuer != expectedIssuer {
		PrintTestError(t, refreshTokenClaims.Issuer, expectedIssuer)
	}

	if len(refreshTokenClaims.Audience) > 0 && refreshTokenClaims.Audience[0] != expectedIssuer {
		PrintTestError(t, refreshTokenClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		PrintTestError(t, err, nil)
	}
}
