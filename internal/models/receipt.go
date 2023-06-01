package models

import (
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/simpleutils"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Receipt struct {
	BaseModel
	Name         string          `gorm:"not null" json:"name"`
	Amount       decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`
	Date         time.Time       `gorm:"not null" json:"date"`
	ResolvedDate *time.Time      `json:"resolvedDate"`
	ImgPath      string          `json:"-"`
	PaidByUserID uint            `json:"paidByUserId"`
	PaidByUser   User            `json:"-"`
	Status       ReceiptStatus   `gorm:"default:'OPEN'; not null" json:"status"`
	GroupId      uint            `gorm:"not null" json:"groupId"`
	Group        Group           `json:"-"`
	Tags         []Tag           `gorm:"many2many:receipt_tags" json:"tags"`
	Categories   []Category      `gorm:"many2many:receipt_categories" json:"categories"`
	ImageFiles   []FileData      `json:"imageFiles"`
	ReceiptItems []Item          `json:"receiptItems"`
	Comments     []Comment       `json:"comments"`
}

func (receiptToUpdate *Receipt) BeforeUpdate(tx *gorm.DB) (err error) {
	if receiptToUpdate.ID > 0 {
		var oldReceipt Receipt

		err := tx.Table("receipts").Where("id = ?", receiptToUpdate.ID).Preload("ImageFiles").Find(&oldReceipt).Error
		if err != nil {
			return err
		}

		if receiptToUpdate.GroupId != oldReceipt.GroupId && len(oldReceipt.ImageFiles) > 0 {
			var oldGroup Group
			var newGroup Group

			err = tx.Table("groups").Where("id = ?", oldReceipt.GroupId).Select("id", "name").Find(&oldGroup).Error
			if err != nil {
				return err
			}

			err = tx.Table("groups").Where("id = ?", receiptToUpdate.GroupId).Select("id", "name").Find(&newGroup).Error
			if err != nil {
				return err
			}

			oldGroupPath, err := simpleutils.BuildGroupPathString(simpleutils.UintToString(oldGroup.ID), oldGroup.Name)
			if err != nil {
				return err
			}

			newGroupPath, err := simpleutils.BuildGroupPathString(simpleutils.UintToString(newGroup.ID), newGroup.Name)
			if err != nil {
				return err
			}

			for _, fileData := range oldReceipt.ImageFiles {
				filename := simpleutils.BuildFileName(simpleutils.UintToString(oldReceipt.ID), simpleutils.UintToString(fileData.ID), fileData.Name)

				oldFilePath := filepath.Join(oldGroupPath, filename)
				newFilePathPath := filepath.Join(newGroupPath, filename)

				err := os.Rename(oldFilePath, newFilePathPath)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func (r *Receipt) AfterUpdate(tx *gorm.DB) (err error) {
	err = tx.Where("receipt_id IS NULL").Delete(&Item{}).Error
	if err != nil {
		return err
	}

	if r.ID > 0 && r.Status == RESOLVED && r.ResolvedDate == nil {
		now := time.Now().UTC()
		err = tx.Table("receipts").Where("id = ?", r.ID).Update("resolved_date", now).Error
	} else if r.ID > 0 && r.Status != RESOLVED && r.ResolvedDate != nil {
		err = tx.Table("receipts").Where("id = ?", r.ID).Update("resolved_date", nil).Error
	}
	if err != nil {
		return err
	}

	return nil
}
