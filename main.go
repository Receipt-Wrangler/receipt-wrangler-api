package main

import (
	connect "receipt-wrangler/api/internal/database"
)

func main() {
	connect.Connect()
	// environment.SetEnv()
	// router := mux.NewRouter()
	// database.ConnectToDatabase()
	// database.MakeMigrations()
	// validator, err := crypto_utils.InitJwtValidator()

	// if err == nil {
	// 	initRoutes(router, validator)
	// 	srv := &http.Server{
	// 		Handler:      router,
	// 		Addr:         "127.0.0.1:8081", // TODO: make configurable
	// 		WriteTimeout: 15 * time.Second,
	// 		ReadTimeout:  15 * time.Second,
	// 	}
	// 	log.Fatal(srv.ListenAndServe())
	// } else {
	// 	log.Fatal(err)
	// }

}

// func initRoutes(router *mux.Router, validator *validator.Validator) {

// }
