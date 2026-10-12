package routers

import (
	"net/http"
	"receipt-wrangler/api/internal/corspolicy"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/mcp"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/oauth"
	"receipt-wrangler/api/internal/services"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
)

func BuildRootRouter() *chi.Mux {
	tokenValidator, err := services.InitTokenValidator()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}
	tokenValidatorMiddleware := jwtmiddleware.New(tokenValidator.ValidateToken)
	env := config.GetDeployEnv()

	rootRouter := chi.NewRouter()

	// TODO: this policy is not ready for production yet. Need to add more configuration options to make sure we aren't using less secure options
	if env == "dev" {
		cors := corspolicy.GetCorsPolicy()
		rootRouter.Use(cors.Handler)
	}

	// Token Refresh Router
	refreshRouter := BuildTokenRefreshRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/token", refreshRouter)

	// Signup Router
	signUpRouter := BuildSignUpRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/signUp", signUpRouter)

	// Login Router
	loginRouter := BuildLoginRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/login", loginRouter)

	// Logout router
	logoutRouter := BuildLogoutRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/logout", logoutRouter)

	// Receipt Router
	receiptRouter := BuildReceiptRouter()
	rootRouter.Mount("/api/receipt", receiptRouter)

	// Receipt Image Router
	receiptImageRouter := BuildReceiptImageRouter()
	rootRouter.Mount("/api/receiptImage", receiptImageRouter)

	// Comment Router
	commentRouter := BuildCommentRouter()
	rootRouter.Mount("/api/comment", commentRouter)

	// Tag Router
	tagRouter := BuildTagRouter()
	rootRouter.Mount("/api/tag", tagRouter)

	// Category Router
	categoryRouter := BuildCategoryRouter()
	rootRouter.Mount("/api/category", categoryRouter)

	// User Router
	userRouter := BuildUserRouter()
	rootRouter.Mount("/api/user", userRouter)

	// Add validaiton on update group that at least one user has owner, and that must have at least 1 user
	// Group Router
	groupRouter := BuildGroupRouter()
	rootRouter.Mount("/api/group", groupRouter)

	// Feature Config Router
	featureConfigRouter := BuildFeatureConfigRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/featureConfig", featureConfigRouter)

	// OIDC relying party (sign in with an external identity provider)
	oidcRouter := BuildOidcRouter()
	rootRouter.Mount("/api/oidc", oidcRouter)

	// OIDC provider administration
	oidcProviderRouter := BuildOidcProviderRouter()
	rootRouter.Mount("/api/oidcProvider", oidcProviderRouter)

	// Migration router
	migrationRouter := chi.NewRouter()
	migrationRouter.Use(middleware.UnifiedAuthMiddleware)
	rootRouter.Mount("/api/migrate", migrationRouter)

	// Search router
	searchRouter := BuildSearchRouter()
	rootRouter.Mount("/api/search", searchRouter)

	// Notification router
	notificationRouter := BuildNotificationRouter()
	rootRouter.Mount("/api/notifications", notificationRouter)

	//User Preferences router
	userPreferencesRouter := BuildUserPreferencesRouter()
	rootRouter.Mount("/api/userPreferences", userPreferencesRouter)

	// Dashboard router
	dashboardRouter := BuildDashboardRouter()
	rootRouter.Mount("/api/dashboard", dashboardRouter)

	// System email router
	systemEmailRouter := BuildSystemEmailRouter()
	rootRouter.Mount("/api/systemEmail", systemEmailRouter)

	// System Task router
	systemTaskRouter := BuildSystemTaskRouter()
	rootRouter.Mount("/api/systemTask", systemTaskRouter)

	// Receipt Processing Settings router
	receiptProcessingSettingsRouter := BuildReceiptProcessingSettingsRouter()
	rootRouter.Mount("/api/receiptProcessingSettings", receiptProcessingSettingsRouter)

	// Prompt router
	promptRouter := BuildPromptRouter()
	rootRouter.Mount("/api/prompt", promptRouter)

	// System Settings router
	systemSettingsRouter := BuildSystemSettingsRouter()
	rootRouter.Mount("/api/systemSettings", systemSettingsRouter)

	// Import router
	importRouter := BuildImportRouter()
	rootRouter.Mount("/api/import", importRouter)

	// Export router
	exportRouter := BuildExportRouter()
	rootRouter.Mount("/api/export", exportRouter)

	// Custom Field router
	customFieldRouter := BuildCustomFieldRouter()
	rootRouter.Mount("/api/customField", customFieldRouter)

	// API Key router
	apiKeyRouter := BuildApiKeyRouter()
	rootRouter.Mount("/api/apiKey", apiKeyRouter)

	// Widget router
	widgetRouter := BuildWidgetRouter()
	rootRouter.Mount("/api/widget", widgetRouter)

	// Report router
	reportRouter := BuildReportRouter()
	rootRouter.Mount("/api/report", reportRouter)

	// Permission router
	permissionRouter := BuildPermissionRouter()
	rootRouter.Mount("/api/permission", permissionRouter)

	// Role router
	roleRouter := BuildRoleRouter()
	rootRouter.Mount("/api/role", roleRouter)

	// MCP server + OAuth 2.1 authorization endpoints. These are mounted at the
	// server root because the OAuth discovery documents must live at well-known
	// root paths. Unlike background workers, HTTP routes can only be mounted
	// once at startup, so rather than re-mounting when the operator toggles the
	// server, the handlers are always mounted and gated at request time by the
	// live MCP_ENABLED System Setting (see mcpEnabledMiddleware).
	mountMcpRoutes(rootRouter)

	return rootRouter
}

// mcpEnabledMiddleware short-circuits MCP/OAuth requests with a 404 whenever the
// MCP server is disabled in System Settings, so the always-mounted routes
// behave as if absent until an admin enables them — no restart required.
func mcpEnabledMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !services.IsMcpEnabled() {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// mountMcpRoutes wires the MCP Streamable HTTP endpoint and the self-hosted
// OAuth 2.1 authorization-server endpoints used by MCP clients such as Claude.
func mountMcpRoutes(rootRouter *chi.Mux) {
	mcpHandler := mcp.NewHandler()

	rootRouter.Group(func(r chi.Router) {
		r.Use(mcpEnabledMiddleware)

		r.Get("/.well-known/oauth-protected-resource", oauth.ProtectedResourceMetadata)
		r.Get("/.well-known/oauth-authorization-server", oauth.AuthorizationServerMetadata)

		r.Post("/oauth/register", oauth.Register)
		r.Get("/oauth/authorize", oauth.AuthorizeForm)
		r.Post("/oauth/authorize", oauth.Authorize)
		r.Post("/oauth/token", oauth.Token)

		r.Handle("/mcp", mcpHandler)
		r.Handle("/mcp/*", mcpHandler)
	})
}
