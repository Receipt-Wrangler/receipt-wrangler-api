package services

import (
	"context"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
)

func customClaims() validator.CustomClaims {
	return &structs.Claims{}
}

// jwtIssuer is the issuer stamped on, and required of, every Receipt Wrangler
// JWT regardless of audience.
const jwtIssuer = "https://receiptWrangler.io"

// defaultAudience is the audience for normal REST API tokens. MCP tokens use a
// distinct, runtime-derived audience instead (see GenerateMcpJWT) so an MCP
// token is rejected everywhere except the MCP endpoints.
const defaultAudience = "https://receiptWrangler.io"

// defaultRefreshTokenLifetime is the fallback refresh-token lifetime, used when
// the System Settings value is unset (0), out of range, or unreadable.
const defaultRefreshTokenLifetime = 24 * time.Hour

// GetRefreshTokenLifetime returns how long a REST refresh token stays valid.
//
// Refresh tokens rotate on every use, so this is an inactivity window rather
// than an absolute session cap: an actively refreshing client is never logged
// out, while an idle one must re-authenticate once it exceeds the window.
func GetRefreshTokenLifetime() time.Duration {
	systemSettings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not read refresh token lifetime, using default: "+err.Error())
		return defaultRefreshTokenLifetime
	}

	return clampRefreshTokenLifetime(systemSettings.RefreshTokenValidForHours)
}

// GetMcpRefreshTokenLifetime returns how long an MCP/OAuth connector refresh
// token stays valid. It is a separate setting from GetRefreshTokenLifetime so a
// long window chosen for human convenience does not silently extend tokens held
// by third-party clients.
func GetMcpRefreshTokenLifetime() time.Duration {
	systemSettings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not read MCP refresh token lifetime, using default: "+err.Error())
		return defaultRefreshTokenLifetime
	}

	return clampRefreshTokenLifetime(systemSettings.McpRefreshTokenValidForHours)
}

// clampRefreshTokenLifetime converts a configured hour count into a duration,
// falling back to the default for anything outside the supported range. This is
// the real safety net — it stops a bad stored value (0, negative, absurd) from
// ever producing a token, independent of whether the value passed command
// validation on the way in.
func clampRefreshTokenLifetime(hours int) time.Duration {
	if hours < commands.MinRefreshTokenValidForHours || hours > commands.MaxRefreshTokenValidForHours {
		return defaultRefreshTokenLifetime
	}

	return time.Duration(hours) * time.Hour
}

func InitTokenValidator() (*validator.Validator, error) {
	return initTokenValidator(defaultAudience)
}

// InitMcpTokenValidator builds a validator that only accepts tokens carrying
// the given MCP audience. Because the normal validator requires
// defaultAudience and our token minting replaces (never appends) the audience,
// an MCP token fails validation on every REST endpoint and a normal token
// fails validation on the MCP endpoints.
func InitMcpTokenValidator(audience string) (*validator.Validator, error) {
	return initTokenValidator(audience)
}

func initTokenValidator(audience string) (*validator.Validator, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		return []byte(config.GetSecretKey()), nil
	}
	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS512,
		jwtIssuer,
		[]string{audience},
		validator.WithCustomClaims(customClaims),
		validator.WithAllowedClockSkew(30*time.Second),
	)

	return jwtValidator, err
}

func LoginUser(loginAttempt commands.LoginCommand) (models.User, bool, error) {
	db := repositories.GetDB()
	firstAdminToLogin := false
	var dbUser models.User

	// Reject empty passwords before any lookup. Dummy/placeholder accounts are
	// stored as bcrypt("") and would otherwise verify against an empty password,
	// and no legitimate login uses an empty password. Centralizing this (and the
	// dummy-user guard below) here means every caller is protected — the REST
	// login handler AND the OAuth/MCP authorize form — rather than relying on
	// each caller to re-check. (REST additionally rejects this at the middleware.)
	if len(loginAttempt.Password) == 0 {
		return models.User{}, false, errors.New("password is required")
	}

	err := db.Model(models.User{}).Where("username = ?", loginAttempt.Username).First(&dbUser).Error
	if err != nil {
		return models.User{}, false, err
	}

	// Dummy (passwordless placeholder) users can never authenticate, regardless
	// of the submitted password.
	if dbUser.IsDummyUser {
		return models.User{}, false, errors.New("dummy users cannot log in")
	}

	err = utils.VerifyPassword(dbUser.Password, loginAttempt.Password)
	if err != nil {
		return models.User{}, false, err
	}

	userRepository := repositories.NewUserRepository(nil)

	// "Administrator" is defined by the app.users.read permission (the modern
	// replacement for the removed UserRole == ADMIN check), resolved from the
	// database rather than the JWT.
	permissionService := NewPermissionService(nil)
	isAdmin, err := permissionService.HasAppPermissions(dbUser.ID, permissions.AppUsersRead)
	if err != nil {
		return models.User{}, false, err
	}
	if isAdmin {
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

// UseSecureCookies reports whether the cookies the API sets should carry the
// Secure attribute. Every cookie builder asks this one function, so the session
// cookies, their expiry cookies and the OIDC binding cookie can never disagree.
//
// It is always true in dev, where BuildTokenCookies uses SameSite=None and
// browsers reject a None cookie without Secure. Everywhere else it is the
// secureCookies System Setting, read live (the IsMcpEnabled pattern) so a save
// applies on the next response. It is off by default because a browser drops a
// Secure cookie set over plain http, which would break login on a LAN install;
// a failed read falls back to that same default.
func UseSecureCookies() bool {
	if config.GetDeployEnv() == "dev" {
		return true
	}

	systemSettings, err := repositories.NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not read secure cookies setting, leaving cookies non-secure: "+err.Error())
		return false
	}

	return systemSettings.SecureCookies
}

func BuildTokenCookies(jwt string, refreshToken string) (http.Cookie, http.Cookie) {
	var sameSite = http.SameSiteStrictMode
	var secure = UseSecureCookies()

	if config.GetDeployEnv() == "dev" {
		sameSite = http.SameSiteNoneMode
	}

	accessTokenCookie := http.Cookie{Name: constants.JwtKey, Value: jwt, HttpOnly: true, Path: "/", Expires: utils.GetAccessTokenExpiryDate().Time, SameSite: sameSite, Secure: secure}
	// Resolved here rather than threaded in from the caller: both call sites
	// (login and token refresh) are REST-only, never MCP, so the app setting is
	// always the right one. Costs one extra System Settings read per login.
	refreshTokenCookie := http.Cookie{Name: constants.RefreshTokenKey, Value: refreshToken, HttpOnly: true, Path: "/", Expires: utils.GetRefreshTokenExpiryDate(GetRefreshTokenLifetime()).Time, SameSite: sameSite, Secure: secure}

	return accessTokenCookie, refreshTokenCookie
}

func PrepareAccessTokenClaims(accessTokenClaims structs.Claims) {
	accessTokenClaims.Issuer = ""
	accessTokenClaims.Audience = make([]string, 0)
}

func GetEmptyAccessTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.JwtKey, Value: "", HttpOnly: false, Path: "/", MaxAge: -1, Secure: UseSecureCookies()}
}

func GetEmptyRefreshTokenCookie() http.Cookie {
	return http.Cookie{Name: constants.RefreshTokenKey, Value: "", HttpOnly: true, Path: "/", MaxAge: -1, Secure: UseSecureCookies()}
}

func GenerateJWT(userId uint) (string, string, structs.Claims, error) {
	return generateTokenPair(userId, defaultAudience, GetRefreshTokenLifetime())
}

// GenerateMcpJWT mints an access + refresh token pair bound to the given MCP
// audience instead of the normal REST audience. The audience is set on BOTH
// tokens: the refresh token matters most, because if it kept the normal
// audience an MCP client could trade it for a full-access token at /api/token.
// Replacing (not appending) the audience ensures the resulting tokens are
// accepted only by the MCP endpoints, which verify this exact audience.
func GenerateMcpJWT(userId uint, audience string) (string, string, structs.Claims, error) {
	return generateTokenPair(userId, audience, GetMcpRefreshTokenLifetime())
}

func generateTokenPair(userId uint, audience string, refreshLifetime time.Duration) (string, string, structs.Claims, error) {
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
		TokenType:          structs.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Audience:  []string{audience},
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
		TokenType:          structs.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Audience:  []string{audience},
			ExpiresAt: utils.GetRefreshTokenExpiryDate(refreshLifetime),
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

	permissionService := NewPermissionService(nil)
	appPermissions, err := permissionService.GetAppPermissionsForUser(userId)
	if err != nil {
		return appData, err
	}

	groupPermissions := make(map[uint][]string, len(groups))
	groupCategories := make(map[uint][]models.Category, len(groups))
	groupTags := make(map[uint][]models.Tag, len(groups))
	// The synthetic "All" group is a real membership where the caller holds an
	// unrestricted role, so resolving its catalog directly would return the whole
	// global pool and leak category/tag names the caller cannot see in any real
	// group. Instead its catalog is the UNION of the caller's per-real-group
	// visible sets, computed after the loop.
	var allGroupIds []uint
	unionCategoryIds := map[uint]struct{}{}
	unionTagIds := map[uint]struct{}{}
	for _, group := range groups {
		perms, err := permissionService.GetGroupPermissionsForUser(userId, group.ID)
		if err != nil {
			return appData, err
		}
		groupPermissions[group.ID] = perms

		if group.IsAllGroup {
			allGroupIds = append(allGroupIds, group.ID)
			continue
		}

		// Per-group category/tag catalog filtered to the caller's grants
		// (full pool when unrestricted). This is how non-admins receive
		// categories/tags now that the flat lists are admin-only.
		visibleCategories, err := permissionService.GetVisibleCategoriesForUser(userId, group.ID, categories)
		if err != nil {
			return appData, err
		}
		groupCategories[group.ID] = visibleCategories
		for _, category := range visibleCategories {
			unionCategoryIds[category.ID] = struct{}{}
		}

		visibleTags, err := permissionService.GetVisibleTagsForUser(userId, group.ID, tags)
		if err != nil {
			return appData, err
		}
		groupTags[group.ID] = visibleTags
		for _, tag := range visibleTags {
			unionTagIds[tag.ID] = struct{}{}
		}
	}

	// Materialize the All-group catalog from the union, preserving the global
	// ordering of categories/tags (which still hold the full pool here; the flat
	// lists are truncated for non-admins below).
	if len(allGroupIds) > 0 {
		unionCategories := make([]models.Category, 0, len(unionCategoryIds))
		for _, category := range categories {
			if _, ok := unionCategoryIds[category.ID]; ok {
				unionCategories = append(unionCategories, category)
			}
		}
		unionTags := make([]models.Tag, 0, len(unionTagIds))
		for _, tag := range tags {
			if _, ok := unionTagIds[tag.ID]; ok {
				unionTags = append(unionTags, tag)
			}
		}
		for _, allGroupId := range allGroupIds {
			groupCategories[allGroupId] = unionCategories
			groupTags[allGroupId] = unionTags
		}
	}

	// The flat global category/tag lists are only for callers who may read the
	// whole pool (app.categories.read / app.tags.read — admins, the category/tag
	// management pages, the role editor). Everyone else receives an empty list
	// and uses the per-group filtered catalogs above.
	canReadAllCategories, err := permissionService.HasAppPermissions(userId, permissions.AppCategoriesRead)
	if err != nil {
		return appData, err
	}
	if !canReadAllCategories {
		categories = []models.Category{}
	}

	canReadAllTags, err := permissionService.HasAppPermissions(userId, permissions.AppTagsRead)
	if err != nil {
		return appData, err
	}
	if !canReadAllTags {
		tags = []models.Tag{}
	}

	// Resolved from the loop's permissions and the groups' preloaded receipt
	// settings, so it costs one query regardless of group count. Must run before
	// the isolation filter below, which only trims members.
	groupReceiptRequirements, err := NewReceiptService(nil).ResolveReceiptRequirementsForGroups(userId, groups, groupPermissions)
	if err != nil {
		return appData, err
	}

	// Member-presence isolation: an isolated member receives only the users and
	// co-members they are allowed to see (no-op for unrestricted viewers). Applied
	// at this serialization boundary, NOT inside GetGroupsForUser / GetAllUserViews,
	// because those feed internal accounting/processing that needs the full roster.
	users, err = permissionService.FilterVisibleUserViews(userId, users)
	if err != nil {
		return appData, err
	}
	if err := permissionService.FilterGroupMembersForGroups(userId, groups); err != nil {
		return appData, err
	}

	// Attach each surviving member's per-member category/tag grants. Loaded after
	// the isolation filter so grants are never fetched for a member the caller
	// cannot see. The fields are `gorm:"-"`, so nothing loads them implicitly.
	if err := repositories.NewGroupMemberRepository(nil).LoadMemberGrantsForGroups(groups); err != nil {
		return appData, err
	}

	// Attach each group's receipt-settings projections — default custom field ids plus the receipt
	// summary configuration. Also `gorm:"-"`, also batched. Must run for EVERY group so an empty set
	// serializes as [] rather than null — the Dart client has no null guard and a null would fail
	// this whole payload on released builds.
	if err := repositories.NewGroupReceiptSettingsRepository(nil).LoadSettingsProjectionsForGroups(groups); err != nil {
		return appData, err
	}

	appData.About = about
	appData.Groups = groups
	appData.Users = users
	appData.UserPreferences = userPreferences
	appData.FeatureConfig = featureConfig
	appData.Categories = categories
	appData.Tags = tags
	appData.GroupCategories = groupCategories
	appData.GroupTags = groupTags
	appData.CurrencyDisplay = systemSettings.CurrencyDisplay
	appData.CurrencyThousandthsSeparator = systemSettings.CurrencyThousandthsSeparator
	appData.CurrencyDecimalSeparator = systemSettings.CurrencyDecimalSeparator
	appData.CurrencySymbolPosition = systemSettings.CurrencySymbolPosition
	appData.CurrencyHideDecimalPlaces = systemSettings.CurrencyHideDecimalPlaces
	appData.TimeZone = repositories.ResolveAppLocation(systemSettings.TimeZone).String()
	appData.Icons = structs.Icons
	appData.AppPermissions = appPermissions
	appData.GroupPermissions = groupPermissions
	appData.GroupReceiptRequirements = groupReceiptRequirements

	if r != nil {
		claims := structs.GetClaims(r)
		PrepareAccessTokenClaims(*claims)
		appData.Claims = *claims
	}

	return appData, nil
}
