package main

import (
	"log"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers/auth"
	auth_middleware "receipt-wrangler/api/internal/middleware/auth"
	auth_utils "receipt-wrangler/api/internal/utils/auth"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.SetConfig()
	db.Connect()
	db.MakeMigrations()

	router := initRoutes()
	serve(router)
}

func serve(router *chi.Mux) {
	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8081", // TODO: make configurable
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func initRoutes() *chi.Mux {
	_, err := auth_utils.InitJwtValidator()
	if err != nil {
		panic(err)
	}

	// tokenRefreshValidator, err := auth_utils.InitTokenRefreshValidator()
	// if err != nil {
	// 	panic(err)
	// }
	// jwtMiddleWare := jwtmiddleware.New(tokenRefreshValidator.ValidateToken)

	rootRouter := chi.NewRouter()

	// Set up token endpoint to refresh our credentials
	// set up jwt validation middleware

	// part of our custom validation will be to compare subject ids, and make sure they are the same, if not, reject
	// also, if we don't get an access token, then we will reject

	// Token Refresh Router
	refreshRouter := chi.NewRouter()
	refreshRouter.Use(auth_middleware.ValidateRefreshToken)
	refreshRouter.Post("/", auth.RefreshToken)

	// Signup Router
	signUpRouter := chi.NewRouter()
	signUpRouter.Use(auth_middleware.SetBodyData)
	signUpRouter.Post("/", auth.SignUp)

	// Login Router
	loginRouter := chi.NewRouter()
	loginRouter.Use(auth_middleware.SetBodyData)
	loginRouter.Post("/", auth.Login)

	rootRouter.Mount("/api/token", refreshRouter)
	rootRouter.Mount("/api/signup", signUpRouter)
	rootRouter.Mount("/api/login", loginRouter)

	return rootRouter
}
