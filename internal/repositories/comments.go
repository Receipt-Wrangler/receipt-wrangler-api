package repositories

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func AddComment(comment models.Comment) (models.Comment, error) {
	db := db.GetDB()

	err := db.Model(&comment).Create(&comment).Error
	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}
