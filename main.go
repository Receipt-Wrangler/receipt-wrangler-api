package main

import (
	"log"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	signUp "receipt-wrangler/api/internal/handlers/auth"
	signUpMiddleware "receipt-wrangler/api/internal/middleware/sign-up"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
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
	signUpRouter.Use(signUpMiddleware.SetBodyData)
	signUpRouter.Post("/", signUp.SignUp)

	rootRouter.Mount("/api/signup", signUpRouter)

	return rootRouter
}
