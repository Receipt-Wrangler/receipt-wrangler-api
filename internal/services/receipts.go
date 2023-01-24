package services

import (
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"strconv"
)

func UserHasAccessToReceipt(userId uint, receiptId string) (bool, error) {
	receipt, err := repositories.GetReceiptById(receiptId)

	if err != nil {
		return false, err
	}

	groupMembers, err := repositories.GetGroupMembersByUserId(userId)
	if err != nil {
		return false, err
	}

	for i := 0; i < len(groupMembers); i++ {
		var groupMember = groupMembers[i]
		if receipt.GroupId == uint(groupMembers[i].GroupID) && (groupMember.GroupRole == models.OWNER || groupMember.GroupRole == models.EDITOR || groupMember.GroupRole == models.VIEWER) {
			return true, nil
		}
	}

	return false, nil
}

func GetReceiptByReceiptImageId(receiptImageId string) (models.Receipt, error) {
	db := db.GetDB()
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Select("receipt_id").First(&fileData).Error
	if err != nil {
		return models.Receipt{}, nil
	}

	receipt, err := repositories.GetReceiptById(strconv.FormatUint(uint64(fileData.ReceiptId), 10))
	if err != nil {
		return models.Receipt{}, nil
	}

	return receipt, nil
}
