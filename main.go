package main

import (
	"log"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	config "receipt-wrangler/api/internal/env"
	login "receipt-wrangler/api/internal/handlers/auth"
	signUp "receipt-wrangler/api/internal/handlers/auth"
	auth_middleware "receipt-wrangler/api/internal/middleware/auth"
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
	rootRouter := chi.NewRouter()

	// Signup Router
	signUpRouter := chi.NewRouter()
	signUpRouter.Use(auth_middleware.SetBodyData)
	signUpRouter.Post("/", signUp.SignUp)

	// Login Router
	loginRouter := chi.NewRouter()
	loginRouter.Use(auth_middleware.SetBodyData)
	loginRouter.Post("/", login.Login)

	rootRouter.Mount("/api/signup", signUpRouter)
	rootRouter.Mount("/api/login", loginRouter)

	return rootRouter
}
