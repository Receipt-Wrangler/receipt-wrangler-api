package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi/v5"
)

// deleteConnectionAs drives DELETE /oidc/connections/{name} with claims built
// the way the named auth source would have built them.
func deleteConnectionAs(t *testing.T, userId uint, providerName string, authSource string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodDelete, "/api/oidc/connections/"+providerName, nil)

	claims := validator.ValidatedClaims{CustomClaims: &structs.Claims{
		UserId:     userId,
		Username:   "connection-owner",
		AuthSource: authSource,
	}}

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("name", providerName)

	ctx := context.WithValue(request.Context(), jwtmiddleware.ContextKey{}, &claims)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeContext)

	recorder := httptest.NewRecorder()
	DeleteOidcConnection(recorder, request.WithContext(ctx))

	return recorder
}

// TestDeleteOidcConnectionRequiresAnInteractiveSession is the other half of the
// rule oidc.LinkStart enforces.
//
// The two endpoints carry the same permission for a stated reason -- adding a
// way to sign into an account is at least as privileged as removing one -- so
// they carry the same credential rule too. Which identity providers may sign
// into an account is for that account's own session to settle, not for a key it
// issued.
func TestDeleteOidcConnectionRequiresAnInteractiveSession(t *testing.T) {
	defer teardownOidcProviderHandlerTest()
	setupOidcProviderHandlerTest(t)

	grantAppPerms(t, oidcTestUserId, permissions.AppAccountUpdate)

	provider := models.OidcProvider{
		Name:         "keycloak",
		DisplayName:  "Keycloak",
		IssuerUrl:    "https://idp.example.com",
		ClientId:     "client",
		ClientSecret: "secret",
		Enabled:      true,
	}
	if err := repositories.GetDB().Create(&provider).Error; err != nil {
		t.Fatalf("failed to create the provider: %v", err)
	}

	identity := models.OidcIdentity{
		OidcProviderId: provider.ID,
		Subject:        "a-subject",
		UserId:         oidcTestUserId,
	}
	if err := repositories.GetDB().Create(&identity).Error; err != nil {
		t.Fatalf("failed to create the identity: %v", err)
	}

	recorder := deleteConnectionAs(t, oidcTestUserId, provider.Name, structs.AuthSourceApiKey)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("expected 403 for an API key, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	// Unmarked claims are denied too -- the check requires an interactive session
	// rather than rejecting a key, so the zero value cannot slip through.
	recorder = deleteConnectionAs(t, oidcTestUserId, provider.Name, "")
	if recorder.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unmarked claims, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var remaining int64
	if err := repositories.GetDB().
		Model(&models.OidcIdentity{}).
		Where("user_id = ?", oidcTestUserId).
		Count(&remaining).Error; err != nil {
		t.Fatalf("failed to count identities: %v", err)
	}

	if remaining != 1 {
		t.Fatalf("a refused unlink removed the identity anyway (%d remaining)", remaining)
	}

	// The positive control: the same call from a real session succeeds, so the
	// 403s above are the credential rule and not some unrelated failure.
	recorder = deleteConnectionAs(t, oidcTestUserId, provider.Name, structs.AuthSourceJwt)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected an interactive session to unlink, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	if err := repositories.GetDB().
		Model(&models.OidcIdentity{}).
		Where("user_id = ?", oidcTestUserId).
		Count(&remaining).Error; err != nil {
		t.Fatalf("failed to count identities: %v", err)
	}

	if remaining != 0 {
		t.Errorf("expected the identity to be removed, %d remaining", remaining)
	}
}
