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

	err = config.SetConfig()
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
	signUpRouter := chi.NewRouter()
	signUpRouter.Use(middleware.SetBodyData)
	signUpRouter.Post("/", handlers.SignUp)
	rootRouter.Mount("/api/signup", signUpRouter)

	// Login Router
	loginRouter := chi.NewRouter()
	loginRouter.Use(middleware.SetBodyData)
	loginRouter.Post("/", handlers.Login)
	rootRouter.Mount("/api/login", loginRouter)

	// Logout router
	logoutRouter := chi.NewRouter()
	logoutRouter.Use(middleware.RevokeRefreshToken)
	logoutRouter.Post("/", handlers.Logout)
	rootRouter.Mount("/api/logout", logoutRouter)

	// Receipt Router
	receiptRouter := chi.NewRouter()
	receiptRouter.Use(tokenValidatorMiddleware.CheckJWT, middleware.SetReceiptBodyData)
	receiptRouter.With(middleware.ValidateGroupAccess).Get("/group/{groupId}", handlers.GetReceiptsForGroup)
	receiptRouter.With(middleware.ValidateReceiptAccess).Get("/{id}", handlers.GetReceipt)
	receiptRouter.With(middleware.ValidateReceiptAccess, middleware.ValidateReceipt).Put("/{id}", handlers.UpdateReceipt)
	receiptRouter.With(middleware.ValidateReceiptAccess).Put("/{id}/toggleIsResolved", handlers.ToggleIsResolved)
	receiptRouter.With(middleware.ValidateReceipt).Post("/", handlers.CreateReceipt)
	receiptRouter.With(middleware.ValidateReceiptAccess).Delete("/{id}", handlers.DeleteReceipt)
	rootRouter.Mount("/api/receipt", receiptRouter)

	// Receipt Image Router
	receiptImageRouter := chi.NewRouter()
	receiptImageRouter.Use(tokenValidatorMiddleware.CheckJWT, middleware.SetReceiptImageData)
	receiptImageRouter.With(middleware.ValidateReceiptImageAccess).Get("/{id}", handlers.GetReceiptImage)
	receiptImageRouter.With(middleware.ValidateReceiptImageAccess).Delete("/{id}", handlers.RemoveReceiptImage)
	receiptImageRouter.Post("/", handlers.UploadReceiptImage)
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
	
	userRouter.With(middleware.SetUserData, middleware.ValidateRole(models.ADMIN)).Post("/{id}", handlers.UpdateUser) // TODO: Add middleware to validate user role
	userRouter.Get("/amountOwedForUser", handlers.GetAmountOwedForUser)
	rootRouter.Mount("/api/user", userRouter)

	// Group Router
	groupRouter := chi.NewRouter()
	groupRouter.Use(tokenValidatorMiddleware.CheckJWT)
	groupRouter.Get("/", handlers.GetGroupsForUser)
	rootRouter.Mount("/api/group", groupRouter)

	return rootRouter
}
