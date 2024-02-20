package services

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func customClaims() validator.CustomClaims {
	return &structs.Claims{}
}

func InitTokenValidator() (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		config := config.GetConfig()
		return []byte(config.SecretKey), nil
	}
	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS512,
		"https://receiptWrangler.io",
		[]string{"https://receiptWrangler.io"},
		validator.WithCustomClaims(customClaims),
		validator.WithAllowedClockSkew(30*time.Second),
	)

	return jwtValidator, err
}

func LoginUser(loginAttempt commands.LoginCommand) (models.User, error) {
	db := repositories.GetDB()
	var dbUser models.User

	err := db.Model(models.User{}).Where("username = ?", loginAttempt.Username).First(&dbUser).Error
	if err != nil {
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginAttempt.Password))
	if err != nil {
		return models.User{}, err
	}

	return dbUser, nil
}

func BuildTokenCookies(jwt string, refreshToken string) (http.Cookie, http.Cookie) {
	accessTokenCookie := http.Cookie{Name: constants.JWT_KEY, Value: jwt, HttpOnly: true, Path: "/", Expires: utils.GetAccessTokenExpiryDate().Time}
	refreshTokenCookie := http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: refreshToken, HttpOnly: true, Path: "/", Expires: utils.GetRefreshTokenExpiryDate().Time}

	return accessTokenCookie, refreshTokenCookie
}

func PrepareAccessTokenClaims(accessTokenClaims structs.Claims) {
	accessTokenClaims.Issuer = ""
	accessTokenClaims.Audience = make([]string, 0)
}

func GetEmptyAccessTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.JWT_KEY, Value: "", HttpOnly: false, Path: "/", MaxAge: -1}
}

func GetEmptyRefreshTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.REFRESH_TOKEN_KEY, Value: "", HttpOnly: true, Path: "/", MaxAge: -1}
}

func GenerateJWT(userId uint) (string, string, structs.Claims, error) {
	db := repositories.GetDB()
	config := config.GetConfig()
	var user models.User

	err := db.Model(models.User{}).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return "", "", structs.Claims{}, err
	}

	accessTokenClaims := structs.Claims{
		DefaultAvatarColor: user.DefaultAvatarColor,
		Displayname:        user.DisplayName,
		UserId:             user.ID,
		Username:           user.Username,
		UserRole:           user.UserRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: utils.GetAccessTokenExpiryDate(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)
	signedString, err := accessToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", structs.Claims{}, err
	}

	refreshTokenClaims := structs.Claims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: utils.GetRefreshTokenExpiryDate(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", "", structs.Claims{}, err
	}

	token := models.RefreshToken{
		UserId: user.ID,
		Token:  refreshTokenString,
		IsUsed: false,
	}

	err = db.Model(&models.RefreshToken{}).Create(&token).Error

	if err != nil {
		return "", "", structs.Claims{}, err
	}

	return signedString, refreshTokenString, accessTokenClaims, nil
}

func GetAppData(userId uint, r *http.Request) (structs.AppData, error) {
	appData := structs.AppData{}
	groupService := NewGroupService(nil)
	userRepository := repositories.NewUserRepository(nil)
	userPreferenceRepository := repositories.NewUserPreferencesRepository(nil)
	categoryRepository := repositories.NewCategoryRepository(nil)
	tagRepository := repositories.NewTagsRepository(nil)
	stringUserId := simpleutils.UintToString(userId)

	groups, err := groupService.GetGroupsForUser(stringUserId)
	if err != nil {
		return appData, err
	}

	users, err := userRepository.GetAllUserViews()
	if err != nil {
		return appData, err
	}

	userPreferences, err := userPreferenceRepository.GetUserPreferencesOrCreate(userId)
	if err != nil {
		return appData, err
	}

	categories, err := categoryRepository.GetAllCategories("*")
	if err != nil {
		return appData, err
	}

	tags, err := tagRepository.GetAllTags("*")
	if err != nil {
		return appData, err
	}

	appData.Groups = groups
	appData.Users = users
	appData.UserPreferences = userPreferences
	appData.FeatureConfig = config.GetFeatureConfig()
	appData.Categories = categories
	appData.Tags = tags

	if r != nil {
		claims := structs.GetJWT(r)
		PrepareAccessTokenClaims(*claims)
		appData.Claims = *claims
	}

	return appData, nil
}
