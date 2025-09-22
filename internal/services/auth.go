package services

import (
	"context"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func customClaims() validator.CustomClaims {
	return &structs.Claims{}
}

func InitTokenValidator() (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		return []byte(config.GetSecretKey()), nil
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

func LoginUser(loginAttempt commands.LoginCommand) (models.User, bool, error) {
	db := repositories.GetDB()
	firstAdminToLogin := false
	var dbUser models.User

	err := db.Model(models.User{}).Where("username = ?", loginAttempt.Username).First(&dbUser).Error
	if err != nil {
		return models.User{}, false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginAttempt.Password))
	if err != nil {
		return models.User{}, false, err
	}

	userRepository := repositories.NewUserRepository(nil)

	if dbUser.UserRole == models.ADMIN {
		firstAdminToLogin, err = userRepository.IsFirstAdminToLogin()
		if err != nil {
			return models.User{}, false, err
		}
	}

	lastLoginDate, err := userRepository.UpdateUserLastLoginDate(dbUser.ID)
	if err != nil {
		return models.User{}, false, err
	}

	dbUser.LastLoginDate = &lastLoginDate
	return dbUser, firstAdminToLogin, nil
}

func BuildTokenCookies(jwt string, refreshToken string) (http.Cookie, http.Cookie) {
	var env = config.GetDeployEnv()
	var sameSite = http.SameSiteStrictMode
	var secure = false

	if env == "dev" {
		sameSite = http.SameSiteNoneMode
		secure = true
	}

	accessTokenCookie := http.Cookie{Name: constants.JwtKey, Value: jwt, HttpOnly: true, Path: "/", Expires: utils.GetAccessTokenExpiryDate().Time, SameSite: sameSite, Secure: secure}
	refreshTokenCookie := http.Cookie{Name: constants.RefreshTokenKey, Value: refreshToken, HttpOnly: true, Path: "/", Expires: utils.GetRefreshTokenExpiryDate().Time, SameSite: sameSite, Secure: secure}

	return accessTokenCookie, refreshTokenCookie
}

func PrepareAccessTokenClaims(accessTokenClaims structs.Claims) {
	accessTokenClaims.Issuer = ""
	accessTokenClaims.Audience = make([]string, 0)
}

func GetEmptyAccessTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.JwtKey, Value: "", HttpOnly: false, Path: "/", MaxAge: -1}
}

func GetEmptyRefreshTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.RefreshTokenKey, Value: "", HttpOnly: true, Path: "/", MaxAge: -1}
}

func GenerateJWT(userId uint) (string, string, structs.Claims, error) {
	db := repositories.GetDB()
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
	signedString, err := accessToken.SignedString([]byte(config.GetSecretKey()))

	if err != nil {
		return "", "", structs.Claims{}, err
	}

	refreshTokenId, err := utils.GetRandomString(16)
	if err != nil {
		return "", "", structs.Claims{}, err
	}

	refreshTokenClaims := structs.Claims{
		DefaultAvatarColor: user.DefaultAvatarColor,
		Displayname:        user.DisplayName,
		UserId:             user.ID,
		Username:           user.Username,
		UserRole:           user.UserRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://receiptWrangler.io",
			Audience:  []string{"https://receiptWrangler.io"},
			ExpiresAt: utils.GetRefreshTokenExpiryDate(),
			ID:        refreshTokenId,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.GetSecretKey()))
	if err != nil {
		return "", "", structs.Claims{}, err
	}

	hashTokenString := utils.Sha256Hash([]byte(refreshTokenString))
	expiresAtFloat := float64(refreshTokenClaims.ExpiresAt.Unix())
	expiresAt := time.Unix(int64(expiresAtFloat), 0).UTC()

	token := models.RefreshToken{
		UserId:    user.ID,
		Token:     hashTokenString,
		IsUsed:    false,
		ExpiresAt: expiresAt,
	}

	err = db.Model(&models.RefreshToken{}).Create(&token).Error
	if err != nil {
		return "", "", structs.Claims{}, err
	}

	return signedString, refreshTokenString, accessTokenClaims, nil
}

func GetAppData(userId uint, r *http.Request) (structs.AppData, error) {
	appData := structs.AppData{}

	aboutRepository := repositories.NewAboutRepository(nil)
	groupService := NewGroupService(nil)
	userRepository := repositories.NewUserRepository(nil)
	userPreferenceRepository := repositories.NewUserPreferencesRepository(nil)
	categoryRepository := repositories.NewCategoryRepository(nil)
	systemSettingsService := NewSystemSettingsService(nil)
	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	tagRepository := repositories.NewTagsRepository(nil)
	stringUserId := utils.UintToString(userId)

	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return appData, err
	}

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

	featureConfig, err := systemSettingsService.GetFeatureConfig()
	if err != nil {
		return appData, err
	}

	about, err := aboutRepository.GetAboutData()
	if err != nil {
		return appData, err
	}

	appData.About = about
	appData.Groups = groups
	appData.Users = users
	appData.UserPreferences = userPreferences
	appData.FeatureConfig = featureConfig
	appData.Categories = categories
	appData.Tags = tags
	appData.CurrencyDisplay = systemSettings.CurrencyDisplay
	appData.CurrencyThousandthsSeparator = systemSettings.CurrencyThousandthsSeparator
	appData.CurrencyDecimalSeparator = systemSettings.CurrencyDecimalSeparator
	appData.CurrencySymbolPosition = systemSettings.CurrencySymbolPosition
	appData.CurrencyHideDecimalPlaces = systemSettings.CurrencyHideDecimalPlaces
	appData.Icons = structs.Icons

	if r != nil {
		claims := structs.GetClaims(r)
		PrepareAccessTokenClaims(*claims)
		appData.Claims = *claims
	}

	return appData, nil
}
