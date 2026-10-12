package oidc

import (
	"errors"
	"net/http"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

// ErrInvalidExchange is returned for every failure mode of the exchange
// endpoint. Unknown, expired, already-used and wrong-verifier are deliberately
// indistinguishable to the caller.
var ErrInvalidExchange = errors.New("invalid or expired exchange code")

// Exchange redeems the mobile handoff code for a real session. Unauthenticated:
// the code plus its PKCE proof IS the authentication.
//
// The response is the full AppData with jwt and refreshToken populated -- exactly
// the shape POST /login/?tokensInBody=true returns -- so the mobile client reuses
// its existing storeAppData path unchanged. No cookie is ever set here.
func Exchange(w http.ResponseWriter, r *http.Request) (structs.AppData, error) {
	command := commands.OidcExchangeCommand{}

	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		return structs.AppData{}, err
	}

	vErr := command.Validate()
	if len(vErr.Errors) > 0 {
		return structs.AppData{}, ErrInvalidExchange
	}

	// Load WITHOUT consuming, so the PKCE proof can be checked first. Burning the
	// code up front would let anyone who intercepted the redirect destroy a valid
	// code out from under the real app just by presenting a wrong verifier.
	row, err := getExchangeCode(command.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return structs.AppData{}, ErrInvalidExchange
		}

		return structs.AppData{}, err
	}

	if !utils.VerifyPkceS256(command.CodeVerifier, row.CodeChallenge) {
		return structs.AppData{}, ErrInvalidExchange
	}

	consumed, err := consumeExchangeCode(command.Code)
	if err != nil {
		return structs.AppData{}, err
	}

	if !consumed {
		return structs.AppData{}, ErrInvalidExchange
	}

	jwt, refreshToken, accessTokenClaims, err := services.GenerateJWT(row.UserId)
	if err != nil {
		return structs.AppData{}, err
	}
	services.PrepareAccessTokenClaims(accessTokenClaims)

	appData, err := services.GetAppData(row.UserId, nil)
	if err != nil {
		return structs.AppData{}, err
	}

	_, err = repositories.NewUserRepository(nil).UpdateUserLastLoginDate(row.UserId)
	if err != nil {
		return structs.AppData{}, err
	}

	appData.Jwt = jwt
	appData.RefreshToken = refreshToken
	appData.Claims = accessTokenClaims

	return appData, nil
}
