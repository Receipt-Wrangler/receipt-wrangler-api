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
	logoutRouter.Use(middleware.RevokeRefreshToken)
	logoutRouter.With(middleware.RevokeRefreshToken).Post("/", handlers.Logout)
	rootRouter.Mount("/api/logout", logoutRouter)

	// Receipt Router
	receiptRouter := chi.NewRouter()
	receiptRouter.Use(tokenValidatorMiddleware.CheckJWT, middleware.SetReceiptBodyData)
	receiptRouter.With(middleware.ValidateGroupRole(models.VIEWER)).Get("/group/{groupId}", handlers.GetReceiptsForGroup)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceipt)
	receiptRouter.With(middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Put("/{id}", handlers.UpdateReceipt)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Put("/{id}/toggleIsResolved", handlers.ToggleIsResolved)
	receiptRouter.With(middleware.ValidateGroupRole(models.EDITOR), middleware.ValidateReceipt).Post("/", handlers.CreateReceipt)
	receiptRouter.With(middleware.SetReceiptGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.DeleteReceipt)
	rootRouter.Mount("/api/receipt", receiptRouter)

	// Receipt Image Router
	receiptImageRouter := chi.NewRouter()
	receiptImageRouter.Use(tokenValidatorMiddleware.CheckJWT)
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.VIEWER)).Get("/{id}", handlers.GetReceiptImage)
	receiptImageRouter.With(middleware.SetReceiptImageGroupId, middleware.ValidateGroupRole(models.EDITOR)).Delete("/{id}", handlers.RemoveReceiptImage)
	receiptImageRouter.With(middleware.SetReceiptImageData, middleware.ValidateGroupRole(models.EDITOR)).Post("/", handlers.UploadReceiptImage)
	rootRouter.Mount("/api/receiptImage", receiptImageRouter)

	// Tag Router
	tagRouter := chi.NewRouter()
	tagRouter.Use(tokenValidatorMiddleware.CheckJWT)
	tagRouter.Get("/", handlers.GetAllTags)
	rootRouter.Mount("/api/tag", tagRouter)

	// Category Router
	categoryRouter := chi.NewRouter()
	categoryRouter.Use(tokenValidatorMiddleware.CheckJWT)
	categoryRouter.Get("/", handlers.GetAllCategories)
	rootRouter.Mount("/api/category", categoryRouter)

	// User Router
	userRouter := chi.NewRouter()
	userRouter.Use(tokenValidatorMiddleware.CheckJWT)
	userRouter.Get("/", handlers.GetAllUsers)
	userRouter.Get("/{username}", handlers.GetUsernameCount)
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN), middleware.ValidateUserData(true)).Post("/", handlers.CreateUser)
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN)).Post("/{id}", handlers.UpdateUser)
	userRouter.With(middleware.SetResetPasswordData, middleware.ValidateRole(models.ADMIN)).Post("/{id}", handlers.ResetPassword)
	userRouter.With(middleware.ValidateRole(models.ADMIN)).Delete("/{id}", handlers.DeleteUser)
	userRouter.With(middleware.ValidateGroupRole(models.VIEWER)).Get("/amountOwedForUser/{groupId}", handlers.GetAmountOwedForUser)
	rootRouter.Mount("/api/user", userRouter)

	// Add validaiton on update group that at least one user has owner, and that must have at least 1 user
	// Group Router
	groupRouter := chi.NewRouter()
	groupRouter.Use(tokenValidatorMiddleware.CheckJWT)
	groupRouter.Get("/", handlers.GetGroupsForUser)
	groupRouter.With(middleware.ValidateGroupRole(models.VIEWER)).Get("/{groupId}", handlers.GetGroupById)
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{})).Post("/", handlers.CreateGroup)
	groupRouter.With(middleware.SetGeneralBodyData("group", models.Group{}), middleware.ValidateGroupRole(models.OWNER)).Put("/{groupId}", handlers.UpdateGroup)
	groupRouter.With(middleware.ValidateGroupRole(models.OWNER), middleware.CanDeleteGroup).Delete("/{groupId}", handlers.DeleteGroup)
	rootRouter.Mount("/api/group", groupRouter)

	// Feature Config Router
	featureConfigRouter := chi.NewRouter()
	featureConfigRouter.Get("/", handlers.GetFeatureConfig)
	rootRouter.Mount("/api/featureConfig", featureConfigRouter)

	return rootRouter
}
