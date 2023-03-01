package services

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

func DeleteUser(userId string) error {
	db := db.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		var receipts []models.Receipt

		// Remove receipt items
		txErr := tx.Where("charged_to_user_id = ?", userId).Delete(&models.Item{}).Error
		if txErr != nil {
			return txErr
		}

		// Remove group members
		txErr = tx.Where("user_id = ?", userId).Delete(&models.GroupMember{}).Error
		if txErr != nil {
			return txErr
		}

		// Remove receipts that the user paid
		txErr = tx.Model(models.Receipt{}).Where("paid_by = ?", userId).Select("id").Find(&receipts).Error
		if txErr != nil {
			return txErr
		}

		for i := 0; i < len(receipts); i++ {
			txErr = DeleteReceipt(utils.UintToString(receipts[i].ID))
			if txErr != nil {
				return txErr
			}
		}

		// Remove user
		txErr = tx.Where("id = ?", userId).Delete(&models.User{}).Error
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
