package main

import (
	"fmt"
	"net/http"
	"os"
	"receipt-wrangler/api/internal/corspolicy"
	"receipt-wrangler/api/internal/email"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/routers"
	"receipt-wrangler/api/internal/services"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func main() {
	err := logging.InitLog()
	if err != nil {
		fmt.Println("Failed to initialize log")
		os.Exit(1)
	}

	logging.LogStd(logging.LOG_LEVEL_INFO, "Initializing...")

	err = config.SetConfigs()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	config.CheckRequiredEnvironmentVariables()

	err = repositories.Connect()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	err = repositories.MakeMigrations()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	err = repositories.InitDB()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	logging.LogStd(logging.LOG_LEVEL_INFO, "Initializing Imagick...")
	imagick.Initialize()
	defer imagick.Terminate()

	err = tryStartEmailPolling()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_FATAL, err.Error())
	}

	router := initRoutes()
	serve(router)
}

func serve(router *chi.Mux) {
	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8081",
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
	}
	logging.LogStd(logging.LOG_LEVEL_INFO, "Initialize completed")
	logging.LogStd(logging.LOG_LEVEL_INFO, "Listening on port 8081")
	logging.LogStd(logging.LOG_LEVEL_FATAL, srv.ListenAndServe())
}

func initRoutes() *chi.Mux {
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
	refreshRouter := routers.BuildTokenRefreshRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/token", refreshRouter)

	// Signup Router
	signUpRouter := routers.BuildSignUpRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/signUp", signUpRouter)

	// Login Router
	loginRouter := routers.BuildLoginRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/login", loginRouter)

	// Logout router
	logoutRouter := routers.BuildLogoutRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/logout", logoutRouter)

	// Receipt Router
	receiptRouter := routers.BuildReceiptRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receipt", receiptRouter)

	// Receipt Image Router
	receiptImageRouter := routers.BuildReceiptImageRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receiptImage", receiptImageRouter)

	// Comment Router
	commentRouter := routers.BuildCommentRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/comment", commentRouter)

	// Tag Router
	tagRouter := routers.BuildTagRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/tag", tagRouter)

	// Category Router
	categoryRouter := routers.BuildCategoryRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/category", categoryRouter)

	// User Router
	userRouter := routers.BuildUserRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/user", userRouter)

	// Add validaiton on update group that at least one user has owner, and that must have at least 1 user
	// Group Router
	groupRouter := routers.BuildGroupRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/group", groupRouter)

	// Feature Config Router
	featureConfigRouter := routers.BuildFeatureConfigRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/featureConfig", featureConfigRouter)

	// Migration router
	migrationRouter := chi.NewRouter()
	migrationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidatorMiddleware.CheckJWT)
	rootRouter.Mount("/api/migrate", migrationRouter)

	// Search router
	searchRouter := routers.BuildSearchRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/search", searchRouter)

	// Notification router
	notificationRouter := routers.BuildNotificationRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/notifications", notificationRouter)

	//User Preferences router
	userPreferencesRouter := routers.BuildUserPreferencesRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/userPreferences", userPreferencesRouter)

	// Dashboard router
	dashboardRouter := routers.BuildDashboardRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/dashboard", dashboardRouter)

	// System email router
	systemEmailRouter := routers.BuildSystemEmailRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemEmail", systemEmailRouter)

	// System Task router
	systemTaskRouter := routers.BuildSystemTaskRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemTask", systemTaskRouter)

	// Receipt Processing Settings router
	receiptProcessingSettingsRouter := routers.BuildReceiptProcessingSettingsRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/receiptProcessingSettings", receiptProcessingSettingsRouter)

	// Prompt router
	promptRouter := routers.BuildPromptRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/prompt", promptRouter)

	// System Settings router
	systemSettingsRouter := routers.BuildSystemSettingsRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/systemSettings", systemSettingsRouter)

	// Import router
	importRouter := routers.BuildImportRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/import", importRouter)

	return rootRouter
}

func tryStartEmailPolling() error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	systemSettingsService := services.NewSystemSettingsService(nil)
	featureConfig, err := systemSettingsService.GetFeatureConfig()
	if err != nil {
		return err
	}

	if systemSettings.EmailPollingInterval > 0 && featureConfig.AiPoweredReceipts {
		err = email.StartEmailPolling()
		if err != nil {
			return err
		}
	}

	return nil
}
