package commands

import (
	"receipt-wrangler/api/internal/models"
	"testing"
)

func validOidcProviderCommand() UpsertOidcProviderCommand {
	secret := "a-client-secret"

	return UpsertOidcProviderCommand{
		Name:         "google",
		DisplayName:  "Google",
		IssuerUrl:    "https://accounts.google.com",
		ClientId:     "a-client-id",
		ClientSecret: &secret,
		Scope:        "openid profile email",
		Enabled:      true,
	}
}

func TestUpsertOidcProviderCommandValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*UpsertOidcProviderCommand)
		isCreate  bool
		errorKey  string
		wantError bool
	}{
		{name: "accepts a valid create", isCreate: true},
		{
			name:      "requires a name",
			mutate:    func(c *UpsertOidcProviderCommand) { c.Name = "" },
			isCreate:  true,
			errorKey:  "name",
			wantError: true,
		},
		{
			name:      "rejects a name with characters unsafe in a URL path",
			mutate:    func(c *UpsertOidcProviderCommand) { c.Name = "My Provider!" },
			isCreate:  true,
			errorKey:  "name",
			wantError: true,
		},
		{
			// Otherwise a provider named "exchange" would be unreachable at its own
			// login URL, which is a confusing way to fail.
			name:      "rejects a reserved name",
			mutate:    func(c *UpsertOidcProviderCommand) { c.Name = "callback" },
			isCreate:  true,
			errorKey:  "name",
			wantError: true,
		},
		{
			name:      "requires a display name",
			mutate:    func(c *UpsertOidcProviderCommand) { c.DisplayName = "" },
			isCreate:  true,
			errorKey:  "displayName",
			wantError: true,
		},
		{
			name:      "requires a client id",
			mutate:    func(c *UpsertOidcProviderCommand) { c.ClientId = "" },
			isCreate:  true,
			errorKey:  "clientId",
			wantError: true,
		},
		{
			name:      "requires an issuer URL",
			mutate:    func(c *UpsertOidcProviderCommand) { c.IssuerUrl = "" },
			isCreate:  true,
			errorKey:  "issuerUrl",
			wantError: true,
		},
		{
			name:      "rejects a relative issuer URL",
			mutate:    func(c *UpsertOidcProviderCommand) { c.IssuerUrl = "/accounts" },
			isCreate:  true,
			errorKey:  "issuerUrl",
			wantError: true,
		},
		{
			// The client secret and the authorization code would otherwise cross the
			// wire in the clear.
			name:      "rejects a plain http issuer on a routable host",
			mutate:    func(c *UpsertOidcProviderCommand) { c.IssuerUrl = "http://accounts.example.com" },
			isCreate:  true,
			errorKey:  "issuerUrl",
			wantError: true,
		},
		{
			name:     "allows a plain http issuer on loopback so a local IdP is testable",
			mutate:   func(c *UpsertOidcProviderCommand) { c.IssuerUrl = "http://localhost:8080/realms/test" },
			isCreate: true,
		},
		{
			name:      "rejects an issuer URL carrying credentials",
			mutate:    func(c *UpsertOidcProviderCommand) { c.IssuerUrl = "https://user:token@accounts.example.com" },
			isCreate:  true,
			errorKey:  "issuerUrl",
			wantError: true,
		},
		{
			// Without openid the provider returns no ID token and the callback has
			// nothing to verify an identity from.
			name:      "requires the openid scope",
			mutate:    func(c *UpsertOidcProviderCommand) { c.Scope = "profile email" },
			isCreate:  true,
			errorKey:  "scope",
			wantError: true,
		},
		{
			name:      "requires a scope",
			mutate:    func(c *UpsertOidcProviderCommand) { c.Scope = "" },
			isCreate:  true,
			errorKey:  "scope",
			wantError: true,
		},
		{
			name:      "requires a client secret on create",
			mutate:    func(c *UpsertOidcProviderCommand) { c.ClientSecret = nil },
			isCreate:  true,
			errorKey:  "clientSecret",
			wantError: true,
		},
		{
			// An omitted key means "keep the stored secret", so the edit form can
			// render the field blank rather than round-tripping the secret.
			name:     "allows an omitted client secret on update",
			mutate:   func(c *UpsertOidcProviderCommand) { c.ClientSecret = nil },
			isCreate: false,
		},
		{
			name: "rejects an explicitly blank client secret",
			mutate: func(c *UpsertOidcProviderCommand) {
				blank := "   "
				c.ClientSecret = &blank
			},
			isCreate:  false,
			errorKey:  "clientSecret",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := validOidcProviderCommand()
			if test.mutate != nil {
				test.mutate(&command)
			}

			vErr := command.Validate(test.isCreate)

			if test.wantError {
				if _, ok := vErr.Errors[test.errorKey]; !ok {
					t.Errorf("expected an error on %q, got %v", test.errorKey, vErr.Errors)
				}
				return
			}

			if len(vErr.Errors) > 0 {
				t.Errorf("expected no errors, got %v", vErr.Errors)
			}
		})
	}
}

// TestUpsertOidcProviderCommandValidateUpdateRejectsARename pins the immutability
// rule: the name is part of the redirect URI already registered with the identity
// provider, so a rename breaks every future login with a mismatch raised at the
// IdP rather than here.
func TestUpsertOidcProviderCommandValidateUpdateRejectsARename(t *testing.T) {
	existing := models.OidcProvider{Name: "google"}

	command := validOidcProviderCommand()
	if vErr := command.ValidateUpdate(existing); len(vErr.Errors) > 0 {
		t.Errorf("expected an unchanged name to pass, got %v", vErr.Errors)
	}

	command.Name = "google-2"
	vErr := command.ValidateUpdate(existing)

	if _, ok := vErr.Errors["name"]; !ok {
		t.Errorf("expected a rename to be rejected, got %v", vErr.Errors)
	}
}
