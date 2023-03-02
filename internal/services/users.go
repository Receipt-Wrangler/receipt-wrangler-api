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
		var groupIdsToNotDelete []uint

		// Remove receipts that the user paid
		txErr := tx.Model(models.Receipt{}).Where("paid_by_user_id = ?", userId).Select("id").Find(&receipts).Error
		if txErr != nil {
			return txErr
		}

		for i := 0; i < len(receipts); i++ {
			txErr = DeleteReceipt(utils.UintToString(receipts[i].ID))
			if txErr != nil {
				return txErr
			}
		}

		// Remove receipt items
		txErr = tx.Where("charged_to_user_id = ?", userId).Delete(&models.Item{}).Error
		if txErr != nil {
			return txErr
		}

		// Remove groups where the user is the only user
		groups, txErr := GetGroupsForUser(userId)
		if txErr != nil {
			return txErr
		}
		for i := 0; i < len(groups); i++ {
			group := groups[i]
			if len(group.GroupMembers) == 1 {
				txErr := DeleteGroup(utils.UintToString(group.ID))
				if txErr != nil {
					return txErr
				} else {
					groupIdsToNotDelete = append(groupIdsToNotDelete, group.ID)
				}
			}
		}

		// Remove other group members
		for i := 0; i < len(groupIdsToNotDelete); i++ {
			txErr = tx.Where("user_id = ? AND group_id = ?", userId, groupIdsToNotDelete[i]).Delete(&models.GroupMember{}).Error
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
