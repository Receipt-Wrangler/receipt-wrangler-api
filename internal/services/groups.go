package services

import (
	"errors"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetGroupsForUser(userId string) ([]models.Group, error) {
	db := repositories.GetDB()
	var groups []models.Group

	groupMemberRepository := repositories.NewGroupMemberRepository(nil)
	groupMembers, err := groupMemberRepository.GetGroupMembersByUserId(userId)
	if err != nil {
		return nil, err
	}

	groupIds := make([]uint, len(groupMembers))
	for i := 0; i < len(groupMembers); i++ {
		groupIds[i] = groupMembers[i].GroupID
	}

	err = db.Model(models.Group{}).Where("id IN ?", groupIds).Preload(clause.Associations).Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func DeleteGroup(groupId string) error {
	db := repositories.GetDB()
	var receipts []models.Receipt

	uintGroupId, err := simpleutils.StringToUint(groupId)
	if err != nil {
		return err
	}

	groupRepository := repositories.NewGroupRepository(nil)
	isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
	if err != nil || isAllGroup {
		return err
	}

	group, err := groupRepository.GetGroupById(groupId, false)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Model(models.Receipt{}).Where("group_id = ?", groupId).Find(&receipts).Error
		if txErr != nil {
			return txErr
		}

		// Delete receipts in group
		for i := 0; i < len(receipts); i++ {
			txErr = DeleteReceipt(simpleutils.UintToString(receipts[i].ID))
			if txErr != nil {
				return txErr
			}
		}

		// Delete group members
		txErr = tx.Where("group_id = ?", groupId).Delete(&models.GroupMember{}).Error
		if txErr != nil {
			return txErr
		}

		// Unset user preferences
		tx.Model(models.UserPrefernces{}).Where("quick_scan_default_group_id = ?", groupId).Update("quick_scan_default_group_id", nil)

		// Delete Group Settings
		txErr = tx.Model(&group.GroupSettings).Select(clause.Associations).Delete(&group.GroupSettings).Error
		if txErr != nil {
			return txErr
		}

		// Delete group
		txErr = tx.Model(&group.GroupSettings).Select(clause.Associations).Delete(&group).Error
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

func ValidateGroupRole(role models.GroupRole, groupId string, userId string) error {
	groupMap := models.BuildGroupMap()

	groupMemberRepository := repositories.NewGroupMemberRepository(nil)
	groupMember, err := groupMemberRepository.GetGroupMemberByUserIdAndGroupId(userId, groupId)
	if err != nil {
		return err
	}

	var hasAccess = groupMap[groupMember.GroupRole] >= groupMap[role]
	if !hasAccess {
		return errors.New("user does not have access to this group")
	}

	return nil
}
