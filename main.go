package main

import (
	"log"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
	"receipt-wrangler/api/internal/utils"
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
	_, err := utils.InitTokenValidator()
	if err != nil {
		panic(err)
	}

	rootRouter := chi.NewRouter()

	// Token Refresh Router
	refreshRouter := chi.NewRouter()
	refreshRouter.Use(middleware.ValidateRefreshToken)
	refreshRouter.Post("/", handlers.RefreshToken)

	// Signup Router
	signUpRouter := chi.NewRouter()
	signUpRouter.Use(middleware.SetBodyData)
	signUpRouter.Post("/", handlers.SignUp)

	// Login Router
	loginRouter := chi.NewRouter()
	loginRouter.Use(middleware.SetBodyData)
	loginRouter.Post("/", handlers.Login)

	rootRouter.Mount("/api/token", refreshRouter)
	rootRouter.Mount("/api/signup", signUpRouter)
	rootRouter.Mount("/api/login", loginRouter)

	return rootRouter
}
