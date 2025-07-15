package services

import (
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"strconv"
	"time"
)

type ItemData struct {
	ID              uint   `gorm:"column:id"`
	CreatedAt       string `gorm:"column:created_at"`
	UpdatedAt       string `gorm:"column:updated_at"`
	CreatedBy       uint   `gorm:"column:created_by"`
	CreatedByString string `gorm:"column:created_by_string"`
	Amount          string `gorm:"column:amount"`
	ChargedToUserId uint   `gorm:"column:charged_to_user_id"`
	IsTaxed         bool   `gorm:"column:is_taxed"`
	Name            string `gorm:"column:name"`
	ReceiptId       uint   `gorm:"column:receipt_id"`
	Status          string `gorm:"column:status"`
}

type ItemCategoryData struct {
	ReceiptId  string
	CategoryId string
}

type ItemTagData struct {
	ReceiptId string
	TagId     string
}

type ShareCategoryData struct {
	ShareId    uint `gorm:"column:share_id"`
	CategoryId uint `gorm:"column:category_id"`
}

type ShareTagData struct {
	ShareId uint `gorm:"column:share_id"`
	TagId   uint `gorm:"column:tag_id"`
}

func MigrateItemsToShares() error {
	db := repositories.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {

		if tx.Migrator().HasTable("items") && tx.Migrator().HasTable("shares") {
			var items []ItemData
			err := tx.Table("items").Find(&items).Error
			if err != nil {
				return errors.New("failed to read data from items table: " + err.Error())
			}

			if len(items) > 0 {
				var shares []models.Share
				for _, item := range items {
					amount, _ := decimal.NewFromString(item.Amount)
					createdAt, _ := time.Parse("2006-01-02 15:04:05", item.CreatedAt)
					updatedAt, _ := time.Parse("2006-01-02 15:04:05", item.UpdatedAt)

					share := models.Share{
						BaseModel: models.BaseModel{
							ID:              item.ID,
							CreatedAt:       createdAt,
							UpdatedAt:       updatedAt,
							CreatedBy:       &item.CreatedBy,
							CreatedByString: item.CreatedByString,
						},
						Amount:          amount,
						ChargedToUserId: item.ChargedToUserId,
						IsTaxed:         item.IsTaxed,
						Name:            item.Name,
						ReceiptId:       item.ReceiptId,
						Status:          models.ShareStatus(item.Status),
					}
					shares = append(shares, share)
				}
				err = tx.Table("shares").CreateInBatches(shares, 100).Error
				if err != nil {
					return errors.New("failed to copy data from items to shares: " + err.Error())
				}
			}
		}

		if tx.Migrator().HasTable("item_categories") && tx.Migrator().HasTable("share_categories") {
			var itemCategories []ItemCategoryData
			err := tx.Table("item_categories").Find(&itemCategories).Error
			if err != nil {
				return errors.New("failed to read data from item_categories table: " + err.Error())
			}

			if len(itemCategories) > 0 {
				var shareCategories []ShareCategoryData
				for _, itemCategory := range itemCategories {
					shareId, _ := strconv.ParseUint(itemCategory.ReceiptId, 10, 32)
					categoryId, _ := strconv.ParseUint(itemCategory.CategoryId, 10, 32)
					shareCategories = append(shareCategories, ShareCategoryData{
						ShareId:    uint(shareId),
						CategoryId: uint(categoryId),
					})
				}
				err = tx.Table("share_categories").Create(&shareCategories).Error
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
				var shareTags []ShareTagData
				for _, itemTag := range itemTags {
					shareId, _ := strconv.ParseUint(itemTag.ReceiptId, 10, 32)
					tagId, _ := strconv.ParseUint(itemTag.TagId, 10, 32)
					shareTags = append(shareTags, ShareTagData{
						ShareId: uint(shareId),
						TagId:   uint(tagId),
					})
				}
				err = tx.Table("share_tags").Create(&shareTags).Error
				if err != nil {
					return errors.New("failed to copy data from item_tags to share_tags: " + err.Error())
				}
			}
		}

		return nil
	})

	return err
}
