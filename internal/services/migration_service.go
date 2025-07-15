package services

import (
	"errors"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

type ItemCategoryData struct {
	ReceiptId  string
	CategoryId string
}

type ItemTagData struct {
	ReceiptId string
	TagId     string
}

func MigrateItemsToShares() error {
	db := repositories.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		if tx.Migrator().HasTable("items") && tx.Migrator().HasTable("shares") {
			var items []models.Share
			err := tx.Table("items").Find(&items).Error
			if err != nil {
				return errors.New("failed to read data from items table: " + err.Error())
			}

			if len(items) > 0 {
				err = tx.Table("shares").CreateInBatches(items, 100).Error
				if err != nil {
					return errors.New("failed to copy data from items to shares: " + err.Error())
				}
			}
		}

		if tx.Migrator().HasTable("item_categories") && tx.Migrator().HasTable("share_categories") {
			var itemCategories []any
			err := tx.Table("item_categories").Find(&itemCategories).Error
			if err != nil {
				return errors.New("failed to read data from item_categories table: " + err.Error())
			}

			if len(itemCategories) > 0 {
				err = tx.Table("share_categories").Create(&itemCategories).Error
				if err != nil {
					return errors.New("failed to copy data from item_categories to share_categories: " + err.Error())
				}
			}
		}

		if tx.Migrator().HasTable("item_tags") && tx.Migrator().HasTable("share_tags") {
			var itemTags []ItemTagData
			err := tx.Table("item_tags").Find(&itemTags).Error
			if err != nil {
				return errors.New("failed to read data from item_tags table: " + err.Error())
			}

			if len(itemTags) > 0 {
				err = tx.Table("share_tags").Create(&itemTags).Error
				if err != nil {
					return errors.New("failed to copy data from item_tags to share_tags: " + err.Error())
				}
			}
		}

		return nil
	})

	return err
}
