package main

import (
	"log"
	"net/http"
	"os"
	"receipt-wrangler/api/internal/corspolicy"
	"receipt-wrangler/api/internal/email"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/routers"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func main() {
	err := logging.InitLog()
	if err != nil {
		log.Fatal(err.Error())
	}

	logger := logging.GetLogger()
	logger.Print("Initializing app...")
	initLoggers()

	err = config.SetConfigs()
	if err != nil {
		logger.Print(err.Error())
		os.Exit(0)
	}

	err = repositories.Connect()
	if err != nil {
		logger.Print(err.Error())
		os.Exit(0)
	}
	repositories.MakeMigrations()

	logger.Print("Initializing Imagick...")
	imagick.Initialize()
	defer imagick.Terminate()

	appConfig := config.GetConfig()

	if appConfig.Features.AiPoweredReceipts && appConfig.AiSettings.AiType == structs.OPEN_AI {
		logger.Print("Initializing OpenAI Client...")
		err = services.InitOpenAIClient()
		if err != nil {
			logger.Print(err.Error())
			os.Exit(0)
		}
	}

	if appConfig.Features.AiPoweredReceipts && appConfig.AiSettings.AiType == structs.GEMINI {
		logger.Print("Initializing Gemini Client...")
		err := services.InitGeminiClient()
		if err != nil {
			logger.Print(err.Error())
			os.Exit(0)
		}
		defer services.GetGeminiClient().Close()
	}

	if config.GetDeployEnv() != "test" {
		userRepository := repositories.NewUserRepository(nil)
		err = userRepository.CreateUserIfNoneExist()
	}

	if len(appConfig.EmailSettings) > 0 && appConfig.Features.AiPoweredReceipts {
		email.PollEmails()
	}

	if err != nil {
		logger.Print(err.Error())
		os.Exit(0)
	}

	router := initRoutes()
	serve(router)
}

func serve(router *chi.Mux) {
	logger := logging.GetLogger()
	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8081",
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
	}
	logger.Print("Initialize completed")
	logger.Fatal(srv.ListenAndServe())
}

func initLoggers() {
	handlers.InitHandlerLogger()
	middleware.InitMiddlewareLogger()
}

func initRoutes() *chi.Mux {
	featureConfig := config.GetFeatureConfig()
	tokenValidator, err := services.InitTokenValidator()
	if err != nil {
		panic(err)
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
	if featureConfig.EnableLocalSignUp {
		signUpRouter := routers.BuildSignUpRouter(tokenValidatorMiddleware)
		rootRouter.Mount("/api/signUp", signUpRouter)
	}

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
	migrationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidatorMiddleware.CheckJWT, middleware.ValidateRole(models.ADMIN))
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

	return rootRouter
}
