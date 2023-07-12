package main

import (
	"log"
	"net/http"
	"os"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/routers"
	"receipt-wrangler/api/internal/utils"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
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

	err = db.Connect()
	if err != nil {
		logger.Print(err.Error())
		os.Exit(0)
	}
	db.MakeMigrations()

	router := initRoutes()
	serve(router)
}

func serve(router *chi.Mux) {
	logger := logging.GetLogger()
	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8081",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
	tokenValidator, err := utils.InitTokenValidator()
	if err != nil {
		panic(err)
	}
	tokenValidatorMiddleware := jwtmiddleware.New(tokenValidator.ValidateToken)

	rootRouter := chi.NewRouter()

	// Token Refresh Router
	refreshRouter := chi.NewRouter()
	refreshRouter.Use(middleware.ValidateRefreshToken, middleware.RevokeRefreshToken)
	refreshRouter.Post("/", handlers.RefreshToken)
	rootRouter.Mount("/api/token", refreshRouter)

	// Signup Router
	if featureConfig.EnableLocalSignUp {
		signUpRouter := chi.NewRouter()
		signUpRouter.Use(middleware.SetBodyData, middleware.ValidateUserData(false))
		signUpRouter.Post("/", handlers.SignUp)
		rootRouter.Mount("/api/signup", signUpRouter)
	}

	// Login Router
	loginRouter := chi.NewRouter()
	loginRouter.With(middleware.SetBodyData, middleware.ValidateLoginData).Post("/", handlers.Login)
	rootRouter.Mount("/api/login", loginRouter)

	// Logout router
	logoutRouter := chi.NewRouter()
	logoutRouter.With(middleware.RevokeRefreshToken).Post("/", handlers.Logout)
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
	featureConfigRouter := chi.NewRouter()
	featureConfigRouter.Get("/", handlers.GetFeatureConfig)
	rootRouter.Mount("/api/featureConfig", featureConfigRouter)

	// Migration router
	migrationRouter := chi.NewRouter()
	migrationRouter.Use(middleware.MoveJWTCookieToHeader, tokenValidatorMiddleware.CheckJWT, middleware.ValidateRole(models.ADMIN))
	migrationRouter.Get("/isResolvedToStatus", handlers.MigratetionMigrateIsResolvedToStatus)
	migrationRouter.Get("/resolveItemsOnResolvedReceipts", handlers.MigrationUpdateReceiptItemStatuses)
	rootRouter.Mount("/api/migrate", migrationRouter)

	// Search router
	searchRouter := routers.BuildSearchRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/search", searchRouter)

	// Notification router
	notificationRouter := routers.BuildNotificationRouter(tokenValidatorMiddleware)
	rootRouter.Mount("/api/notifications", notificationRouter)

	return rootRouter
}
