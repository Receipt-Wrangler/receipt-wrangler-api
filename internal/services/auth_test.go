package services

import (
	"context"
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func TestInitTokenValidatorReturnsValidator(t *testing.T) {
	v, err := InitTokenValidator()

	if v == nil {
		utils.PrintTestError(t, v, "instance of validator")
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestGenerateJWTGeneratesJWTCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Displayname"
	expectedUsername := "Test"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	db := repositories.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		utils.PrintTestError(t, err.Error(), nil)
	}

	jwt, _, _, err := GenerateJWT(user.ID)
	if err != nil {
		utils.PrintTestError(t, jwt, "jwt token")
	}

	rawJwtStruct, err := v.ValidateToken(context.Background(), jwt)
	if err != nil {
		utils.PrintTestError(t, rawJwtStruct, "claim object")
	}

	jwtClaims := rawJwtStruct.(*validator.ValidatedClaims).CustomClaims.(*structs.Claims)

	if jwt == "nil" {
		utils.PrintTestError(t, jwt, "non empty string")
	}

	if jwtClaims.UserId != user.ID {
		utils.PrintTestError(t, jwtClaims.UserId, user.ID)
	}

	if jwtClaims.Displayname != expectedDisplayname {
		utils.PrintTestError(t, jwtClaims.Displayname, expectedDisplayname)
	}

	if jwtClaims.Username != expectedUsername {
		utils.PrintTestError(t, jwtClaims.Username, expectedUsername)
	}

	if jwtClaims.Issuer != expectedIssuer {
		utils.PrintTestError(t, jwtClaims.Issuer, expectedIssuer)
	}

	if len(jwtClaims.Audience) > 0 && jwtClaims.Audience[0] != expectedIssuer {
		utils.PrintTestError(t, jwtClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestGenerateRefreshTokenCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	expectedIssuer := "https://receiptWrangler.io"
	var user models.User

	v, err := InitTokenValidator()

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	db := repositories.GetDB()
	db.Create(&models.User{
		Username:    expectedUsername,
		Password:    "Password",
		DisplayName: expectedDisplayname,
	})

	if db.Where("username = ?", expectedUsername).Select("id").Find(&user).Error != nil {
		utils.PrintTestError(t, err.Error(), nil)
	}

	_, refreshToken, _, err := GenerateJWT(user.ID)
	if err != nil {
		utils.PrintTestError(t, refreshToken, "refresh token")
	}

	rawRefreshTokenClaims, err := v.ValidateToken(context.Background(), refreshToken)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
		return
	}

	if rawRefreshTokenClaims == nil {
		utils.PrintTestError(t, rawRefreshTokenClaims, "non-nil claim object")
		return
	}

	refreshTokenClaims := rawRefreshTokenClaims.(*validator.ValidatedClaims).CustomClaims.(*structs.Claims)

	if refreshToken == "nil" {
		utils.PrintTestError(t, refreshToken, "non empty string")
	}

	if refreshTokenClaims.UserId != user.ID {
		utils.PrintTestError(t, refreshTokenClaims.UserId, user.ID)
	}

	if refreshTokenClaims.Issuer != expectedIssuer {
		utils.PrintTestError(t, refreshTokenClaims.Issuer, expectedIssuer)
	}

	if len(refreshTokenClaims.Audience) > 0 && refreshTokenClaims.Audience[0] != expectedIssuer {
		utils.PrintTestError(t, refreshTokenClaims.Audience, fmt.Sprintf("[%s]", expectedIssuer))
	}

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
}

func TestShouldLogInUserCorrectly(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	password := "Password"

	userRepository := repositories.NewUserRepository(nil)

	_, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    expectedUsername,
		Password:    password,
		DisplayName: expectedDisplayname,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	user, firstAdminToLogin, err := LoginUser(commands.LoginCommand{
		Username: expectedUsername,
		Password: password,
	})

	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if firstAdminToLogin != true {
		utils.PrintTestError(t, firstAdminToLogin, true)
	}

	if user.LastLoginDate == nil {
		utils.PrintTestError(t, user.LastLoginDate, nil)
	}
}

func TestShouldNotLogUserInWithWrongPassword(t *testing.T) {
	defer repositories.TruncateTestDb()
	expectedDisplayname := "Another displayname"
	expectedUsername := "Another username"
	password := "Password"

	userRepository := repositories.NewUserRepository(nil)

	_, err := userRepository.CreateUser(commands.SignUpCommand{
		Username:    expectedUsername,
		Password:    password,
		DisplayName: expectedDisplayname,
	})
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	_, _, err = LoginUser(commands.LoginCommand{
		Username: expectedUsername,
		Password: "wrong password",
	})

	if err == nil {
		utils.PrintTestError(t, err, "login error")
	}
}
