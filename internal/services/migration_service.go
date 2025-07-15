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
	ID              uint `gorm:"primarykey"`
	CreatedAt       string
	UpdatedAt       string
	CreatedBy       *uint
	CreatedByString string
	Amount          string
	ChargedToUserId uint
	IsTaxed         bool
	Name            string
	ReceiptId       uint
	Status          string
}

type ItemCategoryData struct {
	ReceiptId  string
	CategoryId string
}

type ItemTagData struct {
	ReceiptId string
	TagId     string
}

func parseStringToUint(s string) uint {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(val)
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
					createdAt, _ := time.Parse("2006-01-02 15:04:05", item.CreatedAt)
					updatedAt, _ := time.Parse("2006-01-02 15:04:05", item.UpdatedAt)
					amount, _ := decimal.NewFromString(item.Amount)

					share := models.Share{
						BaseModel: models.BaseModel{
							ID:              item.ID,
							CreatedAt:       createdAt,
							UpdatedAt:       updatedAt,
							CreatedBy:       item.CreatedBy,
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

				err = tx.Table("shares").CreateInBatches(shares, 1000).Error
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
				for _, itemCat := range itemCategories {
					itemId := parseStringToUint(itemCat.ReceiptId)
					categoryId := parseStringToUint(itemCat.CategoryId)

					shareCategories := struct {
						ShareId    uint
						CategoryId uint
					}{
						ShareId:    itemId,
						CategoryId: categoryId,
					}
					err = tx.Table("share_categories").Create(&shareCategories).Error
					if err != nil {
						return errors.New("failed to copy data from item_categories to share_categories: " + err.Error())
					}
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
				for _, itemTag := range itemTags {
					itemId := parseStringToUint(itemTag.ReceiptId)
					tagId := parseStringToUint(itemTag.TagId)

					shareTags := struct {
						ShareId uint
						TagId   uint
					}{
						ShareId: itemId,
						TagId:   tagId,
					}
					err = tx.Table("share_tags").Create(&shareTags).Error
					if err != nil {
						return errors.New("failed to copy data from item_tags to share_tags: " + err.Error())
					}
				}
			}
		}

		return nil
	})

	return err
}
