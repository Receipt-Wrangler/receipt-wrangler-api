package commands

import (
	"encoding/json"
	"net/http"
	"net/url"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"regexp"
	"slices"
	"strings"
)

// ReservedOidcProviderNames are slugs that would shadow a route under
// /api/oidc/. The router keeps its static segments first so resolution is
// unambiguous either way, but a provider named "exchange" would still be
// unreachable at its own login URL, which is a confusing way to fail.
var ReservedOidcProviderNames = []string{
	"login",
	"callback",
	"link",
	"exchange",
	"connections",
}

// OidcRequiredScope is the scope that makes a request an OIDC request rather
// than a plain OAuth one. Without it the provider returns no ID token and the
// callback has nothing to verify an identity from.
const OidcRequiredScope = "openid"

// oidcProviderNameRegex constrains the slug to what is safe in a URL path
// segment: lowercase alphanumerics and dashes, starting and ending with an
// alphanumeric.
var oidcProviderNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)

type UpsertOidcProviderCommand struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	IssuerUrl   string `json:"issuerUrl"`
	ClientId    string `json:"clientId"`
	// ClientSecret is a pointer so an omitted key ("keep the stored secret") is
	// distinguishable from an explicit empty string. The edit form deliberately
	// renders the field blank rather than round-tripping the secret to the client,
	// so an update that does not re-type it must not clear it.
	ClientSecret      *string `json:"clientSecret"`
	Scope             string  `json:"scope"`
	AllowProvisioning bool    `json:"allowProvisioning"`
	LinkByUsername    bool    `json:"linkByUsername"`
	Enabled           bool    `json:"enabled"`
}

func (command *UpsertOidcProviderCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	command.Name = strings.ToLower(strings.TrimSpace(command.Name))
	command.DisplayName = strings.TrimSpace(command.DisplayName)
	command.IssuerUrl = strings.TrimSpace(command.IssuerUrl)
	command.ClientId = strings.TrimSpace(command.ClientId)
	command.Scope = strings.TrimSpace(command.Scope)

	return nil
}

func (command *UpsertOidcProviderCommand) Validate(isCreate bool) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(command.Name) == 0 {
		errors["name"] = "Name is required"
	} else if !oidcProviderNameRegex.MatchString(command.Name) {
		errors["name"] = "Name must be lowercase letters, numbers and dashes, and start and end with a letter or number"
	} else if slices.Contains(ReservedOidcProviderNames, command.Name) {
		errors["name"] = "Name is reserved, pick another"
	}

	if len(command.DisplayName) == 0 {
		errors["displayName"] = "Display name is required"
	}

	if len(command.ClientId) == 0 {
		errors["clientId"] = "Client ID is required"
	}

	if msg := validateOidcIssuerUrl(command.IssuerUrl); len(msg) > 0 {
		errors["issuerUrl"] = msg
	}

	if len(command.Scope) == 0 {
		errors["scope"] = "Scope is required"
	} else if !slices.Contains(strings.Fields(command.Scope), OidcRequiredScope) {
		errors["scope"] = "Scope must include " + OidcRequiredScope
	}

	// On create the secret is mandatory. On update a nil pointer means "keep the
	// stored one", but an explicitly supplied blank is a mistake worth catching --
	// it would leave the provider unable to complete a token exchange.
	if command.ClientSecret == nil {
		if isCreate {
			errors["clientSecret"] = "Client secret is required"
		}
	} else if len(strings.TrimSpace(*command.ClientSecret)) == 0 {
		errors["clientSecret"] = "Client secret cannot be blank"
	}

	vErr.Errors = errors
	return vErr
}

// ValidateUpdate holds the rules that need the currently-stored provider, kept
// beside the create rules rather than in the handler. It is pure.
func (command *UpsertOidcProviderCommand) ValidateUpdate(existing models.OidcProvider) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	// The name is part of the redirect URI already registered at the identity
	// provider, so renaming it would silently break every subsequent login with a
	// redirect_uri mismatch raised at the IdP rather than here.
	if command.Name != existing.Name {
		errors["name"] = "Name cannot be changed"
	}

	vErr.Errors = errors
	return vErr
}

// validateOidcIssuerUrl requires an absolute origin and, outside loopback,
// https. An http issuer on a routable host would send the client secret and the
// authorization code over the wire in the clear; loopback stays allowed so a
// locally-run IdP is testable.
func validateOidcIssuerUrl(raw string) string {
	if len(raw) == 0 {
		return "Issuer URL is required"
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return "Issuer URL must be an absolute URL like https://accounts.google.com"
	}

	if parsed.Scheme == "https" {
		return ""
	}

	if parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname()) {
		return ""
	}

	return "Issuer URL must use https (http is only allowed for localhost)"
}

func isLoopbackHost(hostname string) bool {
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}
