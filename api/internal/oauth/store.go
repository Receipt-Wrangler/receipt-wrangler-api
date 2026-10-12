package oauth

import (
	"encoding/json"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"time"
)

// createClient persists a dynamically registered OAuth client and returns it.
func createClient(clientName string, redirectUris []string) (models.OAuthClient, error) {
	clientId, err := utils.GetRandomUrlSafeString(24)
	if err != nil {
		return models.OAuthClient{}, err
	}

	encodedUris, err := json.Marshal(redirectUris)
	if err != nil {
		return models.OAuthClient{}, err
	}

	client := models.OAuthClient{
		ClientId:     clientId,
		ClientName:   clientName,
		RedirectUris: string(encodedUris),
	}

	if err := repositories.GetDB().Create(&client).Error; err != nil {
		return models.OAuthClient{}, err
	}

	return client, nil
}

// getClient loads a registered client by id. A missing client is reported as
// gorm.ErrRecordNotFound so callers can distinguish "unknown client" from a
// database failure.
func getClient(clientId string) (models.OAuthClient, error) {
	var client models.OAuthClient
	err := repositories.GetDB().Where("client_id = ?", clientId).First(&client).Error
	return client, err
}

// clientAllowsRedirect reports whether redirectUri exactly matches one of the
// client's registered redirect URIs. Exact match (no prefix/substring) is
// required by OAuth 2.1 to prevent open-redirect attacks.
func clientAllowsRedirect(client models.OAuthClient, redirectUri string) bool {
	var registered []string
	if err := json.Unmarshal([]byte(client.RedirectUris), &registered); err != nil {
		return false
	}

	for _, uri := range registered {
		if uri == redirectUri {
			return true
		}
	}

	return false
}

// createAuthorizationCode issues and persists a single-use authorization code
// bound to the authenticated user, client, redirect URI, PKCE challenge, and
// requested resource.
func createAuthorizationCode(
	clientId string,
	userId uint,
	redirectUri string,
	codeChallenge string,
	scope string,
	resource string,
) (string, error) {
	code, err := utils.GetRandomUrlSafeString(32)
	if err != nil {
		return "", err
	}

	authCode := models.OAuthAuthorizationCode{
		Code:          code,
		ClientId:      clientId,
		UserId:        userId,
		RedirectUri:   redirectUri,
		CodeChallenge: codeChallenge,
		Scope:         scope,
		Resource:      resource,
		ExpiresAt:     time.Now().Add(authCodeTTLSeconds * time.Second),
	}

	if err := repositories.GetDB().Create(&authCode).Error; err != nil {
		return "", err
	}

	return code, nil
}

// getAuthorizationCode loads an authorization code without consuming it, so the
// caller can validate the client binding and PKCE before the code is burned. A
// missing code is reported as gorm.ErrRecordNotFound.
func getAuthorizationCode(code string) (models.OAuthAuthorizationCode, error) {
	var authCode models.OAuthAuthorizationCode
	err := repositories.GetDB().Where("code = ?", code).First(&authCode).Error
	return authCode, err
}

// markAuthorizationCodeUsed atomically marks an unused, unexpired code as used
// and reports whether it succeeded. The used/expiry checks live in the UPDATE's
// WHERE clause so that check-and-set is a single atomic statement: concurrent
// redemptions of the same code can never both observe it as unused, and a
// replayed code affects zero rows.
func markAuthorizationCodeUsed(code string) (bool, error) {
	result := repositories.GetDB().
		Model(&models.OAuthAuthorizationCode{}).
		Where("code = ? AND used = ? AND expires_at > ?", code, false, time.Now()).
		Update("used", true)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}
