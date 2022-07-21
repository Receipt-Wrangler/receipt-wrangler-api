package main

import (
	db "receipt-wrangler/api/internal/database"
)

func main() {
	db.Connect()
	db.MakeMigrations()
}
