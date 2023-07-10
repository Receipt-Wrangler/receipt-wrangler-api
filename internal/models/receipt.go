package models

import (
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/simpleutils"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Receipt
//
// swagger:model
type Receipt struct {
	BaseModel

	// Receipt name
	//
	// required: true
	Name string `gorm:"not null" json:"name"`

	// Reciept total amount
	//
	// required: true
	Amount decimal.Decimal `gorm:"not null" json:"amount" sql:"type:decimal(20,3);"`

	// Receipt date
	//
	// required: true
	Date time.Time `gorm:"not null" json:"date"`

	// Date resolved
	//
	// required: false
	ResolvedDate *time.Time `json:"resolvedDate"`

	// User paid foreign key
	//
	// required: true
	PaidByUserID uint `json:"paidByUserId"`

	PaidByUser User `json:"-"`

	// Receipt status
	//
	// required: ture
	Status ReceiptStatus `gorm:"default:'OPEN'; not null" json:"status"`

	// Group foreign key
	//
	// required: true
	GroupId uint `gorm:"not null" json:"groupId"`

	Group Group `json:"-"`

	// Tags associated to receipt
	Tags []Tag `gorm:"many2many:receipt_tags" json:"tags"`

	// Categories associated to receipt
	Categories []Category `gorm:"many2many:receipt_categories" json:"categories"`

	// Files associated to receipt
	ImageFiles []FileData `json:"imageFiles"`

	// Items associated to receipt
	ReceiptItems []Item `json:"receiptItems"`

	// Comments associated to receipt
	Comments []Comment `json:"comments"`
}

func (receiptToUpdate *Receipt) BeforeUpdate(tx *gorm.DB) (err error) {
	if receiptToUpdate.ID > 0 {
		var oldReceipt Receipt

		err := tx.Table("receipts").Where("id = ?", receiptToUpdate.ID).Preload("ImageFiles").Find(&oldReceipt).Error
		if err != nil {
			return err
		}

		if receiptToUpdate.GroupId > 0 && receiptToUpdate.GroupId != oldReceipt.GroupId && len(oldReceipt.ImageFiles) > 0 {
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

	if r.Status == RESOLVED && r.ID > 0 {
		err := updateItemsToResolved(tx, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateItemsToResolved(tx *gorm.DB, r *Receipt) error {
	var items []Item
	var itemIdsToUpdate []uint

	err := tx.Table("items").Where("receipt_id = ?", r.ID).Find(&items).Error
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Status != ITEM_RESOLVED {
			itemIdsToUpdate = append(itemIdsToUpdate, item.ID)
		}
	}

	if len(itemIdsToUpdate) > 0 {
		err := tx.Table("items").Where("id IN ?", itemIdsToUpdate).UpdateColumn("status", ITEM_RESOLVED).Error
		if err != nil {
			return err
		}
	}

	return nil
}
