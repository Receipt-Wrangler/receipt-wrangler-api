package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/corspolicy"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/services"
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
	receiptRouter := BuildReceiptRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receipt", receiptRouter)

	// Receipt Image Router
	receiptImageRouter := BuildReceiptImageRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receiptImage", receiptImageRouter)

	// Comment Router
	commentRouter := BuildCommentRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/comment", commentRouter)

	// Tag Router
	tagRouter := BuildTagRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/tag", tagRouter)

	// Category Router
	categoryRouter := BuildCategoryRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/category", categoryRouter)

	// User Router
	userRouter := BuildUserRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/user", userRouter)

	// Add validaiton on update group that at least one user has owner, and that must have at least 1 user
	// Group Router
	groupRouter := BuildGroupRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/group", groupRouter)

	// Feature Config Router
	featureConfigRouter := BuildFeatureConfigRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/featureConfig", featureConfigRouter)

	// Migration router
	migrationRouter := chi.NewRouter()
	migrationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidatorMiddleware.CheckJWT)
	rootRouter.Mount("/api/migrate", migrationRouter)

	// Search router
	searchRouter := BuildSearchRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/search", searchRouter)

	// Notification router
	notificationRouter := BuildNotificationRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/notifications", notificationRouter)

	//User Preferences router
	userPreferencesRouter := BuildUserPreferencesRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/userPreferences", userPreferencesRouter)

	// Dashboard router
	dashboardRouter := BuildDashboardRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/dashboard", dashboardRouter)

	// System email router
	systemEmailRouter := BuildSystemEmailRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemEmail", systemEmailRouter)

	// System Task router
	systemTaskRouter := BuildSystemTaskRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemTask", systemTaskRouter)

	// Receipt Processing Settings router
	receiptProcessingSettingsRouter := BuildReceiptProcessingSettingsRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receiptProcessingSettings", receiptProcessingSettingsRouter)

	// Prompt router
	promptRouter := BuildPromptRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/prompt", promptRouter)

	// System Settings router
	systemSettingsRouter := BuildSystemSettingsRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemSettings", systemSettingsRouter)

	// Import router
	importRouter := BuildImportRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/import", importRouter)

	// Export router
	exportRouter := BuildExportRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/export", exportRouter)

	return rootRouter
}
