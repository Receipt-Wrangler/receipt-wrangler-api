package services

import (
	db "receipt-wrangler/api/internal/database"

	"gorm.io/gorm"
)

func DeleteUser(userId string) error {
	db := db.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		// Remove receipt items
		// Remove group member
		// Remove receipts that the user paid
		// Remove user
		return nil
	})
	if err != nil {
		return err
	}

	return err
}
