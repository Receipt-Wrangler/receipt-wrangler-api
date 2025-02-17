package services

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

type GroupService struct {
	BaseService
}

func NewGroupService(tx *gorm.DB) GroupService {
	service := GroupService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service GroupService) GetGroupIdsForUser(userId string) ([]uint, error) {
	groupMemberRepository := repositories.NewGroupMemberRepository(nil)
	groupMembers, err := groupMemberRepository.GetGroupMembersByUserId(userId)
	if err != nil {
		return nil, err
	}

	groupIds := make([]uint, len(groupMembers))
	for i := 0; i < len(groupMembers); i++ {
		groupIds[i] = groupMembers[i].GroupID
	}

	return groupIds, nil
}

func (service GroupService) GetGroupsForUser(userId string) ([]models.Group, error) {
	db := service.GetDB()
	var groups []models.Group

	groupIds, err := service.GetGroupIdsForUser(userId)
	if err != nil {
		return nil, err
	}

	err = db.Model(models.Group{}).
		Where("id IN ?", groupIds).
		Preload(clause.Associations).
		Order("is_all_group desc").
		Find(&groups).Error

	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (service GroupService) DeleteGroup(groupId string, allowAllGroupDelete bool) error {
	db := service.GetDB()
	var receipts []models.Receipt

	uintGroupId, err := utils.StringToUint(groupId)
	if err != nil {
		return err
	}

	groupRepository := repositories.NewGroupRepository(nil)

	if !allowAllGroupDelete {
		isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
		if err != nil || isAllGroup {
			return err
		}
	}

	group, err := groupRepository.GetGroupById(groupId, false, false, false)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptService := NewReceiptService(tx)

		txErr := tx.Model(models.Receipt{}).Where("group_id = ?", groupId).Find(&receipts).Error
		if txErr != nil {
			return txErr
		}

		// Delete receipts in group
		for i := 0; i < len(receipts); i++ {
			txErr = receiptService.DeleteReceipt(utils.UintToString(receipts[i].ID))
			if txErr != nil {
				return txErr
			}
		}

		// Delete dashboards in group
		dashboardRepository := repositories.NewDashboardRepository(tx)
		groupDashboards, txErr := dashboardRepository.GetDashboardsByGroupId(group.ID)
		if txErr != nil {
			return txErr
		}

		for _, dashboard := range groupDashboards {
			txErr = dashboardRepository.DeleteDashboardById(dashboard.ID)
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
		if group.GroupSettings.GroupId > 0 {
			txErr = tx.Model(&group.GroupSettings).Select(clause.Associations).Delete(&group.GroupSettings).Error
			if txErr != nil {
				return txErr
			}
		}

		// Delete Group Receipt Settings
		if group.GroupReceiptSettings.GroupId > 0 {
			txErr = tx.Model(&group.GroupReceiptSettings).Select(clause.Associations).Delete(&group.GroupReceiptSettings).Error
			if txErr != nil {
				return txErr
			}
		}

		// Delete group
		txErr = tx.Delete(&group).Error
		if txErr != nil {
			return txErr
		}

		// Remove group's directory
		groupPath, txErr := utils.BuildGroupPathString(utils.UintToString(group.ID), group.Name)
		if txErr != nil {
			return txErr
		}

		txErr = os.Remove(groupPath)
		if txErr != nil {
			logging.LogStd(logging.LOG_LEVEL_INFO, txErr.Error())
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (service GroupService) ValidateGroupRole(role models.GroupRole, groupId string, userId string) error {
	groupMap := models.BuildGroupMap()

	groupMemberRepository := repositories.NewGroupMemberRepository(service.TX)
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
