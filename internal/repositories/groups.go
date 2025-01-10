package repositories

import (
	"errors"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
)

type GroupRepository struct {
	BaseRepository
}

func NewGroupRepository(tx *gorm.DB) GroupRepository {
	repository := GroupRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository GroupRepository) GetPagedGroups(command commands.PagedGroupRequestCommand, userId string) ([]models.Group, int64, error) {
	db := repository.GetDB()
	var results []models.Group
	var count int64

	query := db.Model(&models.Group{}).Where("(is_all_group = ? OR is_all_group IS NULL)", false)

	if !repository.isValidColumn(command.OrderBy) {
		return nil, 0, errors.New("invalid column")
	}

	// Apply filter and set counts
	if command.GroupFilter.AssociatedGroup == commands.ASSOCIATED_GROUP_ALL {
		query.Count(&count)
	} else if command.GroupFilter.AssociatedGroup == commands.ASSOCIATED_GROUP_MINE {
		groupMemberRepository := NewGroupMemberRepository(nil)
		groupMembers, err := groupMemberRepository.GetGroupMembersByUserId(userId)
		if err != nil {
			return nil, 0, err
		}

		groupIds := make([]uint, len(groupMembers))
		for i := 0; i < len(groupMembers); i++ {
			groupIds[i] = groupMembers[i].GroupID
		}

		query = query.Where("id IN ?", groupIds)
		err = query.Count(&count).Error
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply sorting and pagination
	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := query.Preload(clause.Associations).
		Find(&results).
		Error
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (repository GroupRepository) isValidColumn(orderBy string) bool {
	return orderBy == "name" ||
		orderBy == "num_of_members" ||
		orderBy == "default_group" ||
		orderBy == "created_at" ||
		orderBy == "updated_at"
}

func (repository GroupRepository) CreateGroup(command commands.UpsertGroupCommand, userId uint) (models.Group, error) {
	// TODO: move hooks on delete to repository func
	db := repository.GetDB()
	var returnGroup models.Group
	var groupToCreate models.Group

	groupToCreate.Name = command.Name
	groupToCreate.Status = command.Status
	groupToCreate.IsAllGroup = command.IsAllGroup
	for i := 0; i < len(command.GroupMembers); i++ {
		groupMemberCommand := command.GroupMembers[i]
		groupMember := models.GroupMember{
			UserID:    groupMemberCommand.UserID,
			GroupID:   groupMemberCommand.GroupID,
			GroupRole: groupMemberCommand.GroupRole,
		}

		groupToCreate.GroupMembers = append(groupToCreate.GroupMembers, groupMember)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)
		groupSettingsRepository := NewGroupSettingsRepository(tx)
		groupReceiptSettingsRepository := NewGroupReceiptSettingsRepository(tx)

		txErr := tx.Model(&groupToCreate).Create(&groupToCreate).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupMember := models.GroupMember{
			UserID:    userId,
			GroupID:   groupToCreate.ID,
			GroupRole: models.OWNER,
		}

		txErr = tx.Model(&groupMember).Create(&groupMember).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupSettings := models.GroupSettings{
			GroupId: groupToCreate.ID,
		}

		_, txErr = groupSettingsRepository.CreateGroupSettings(groupSettings)
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		_, txErr = groupReceiptSettingsRepository.CreateGroupReceiptSettings(groupToCreate.ID)
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		repository.ClearTransaction()
		return nil
	})
	if err != nil {
		return models.Group{}, err
	}

	err = repository.GetDB().Model(models.Group{}).Where("id = ?", groupToCreate.ID).Preload("GroupMembers").Find(&returnGroup).Error
	if err != nil {
		return models.Group{}, err
	}

	return returnGroup, nil
}

func (repository GroupRepository) UpdateGroup(command commands.UpsertGroupCommand, groupId string) (models.Group, error) {
	// TODO: move hooks from model to repository func
	db := repository.GetDB()

	u64Id, err := utils.StringToUint64(groupId)
	if err != nil {
		return models.Group{}, err
	}

	groupToUpdate := models.Group{
		Name:   command.Name,
		Status: command.Status,
	}
	groupToUpdate.ID = uint(u64Id)

	for i := 0; i < len(command.GroupMembers); i++ {
		groupMemberCommand := command.GroupMembers[i]
		groupMember := models.GroupMember{
			UserID:    groupMemberCommand.UserID,
			GroupID:   groupMemberCommand.GroupID,
			GroupRole: groupMemberCommand.GroupRole,
		}
		groupToUpdate.GroupMembers = append(groupToUpdate.GroupMembers, groupMember)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&groupToUpdate).Omit("ID", "is_all_group").Updates(&groupToUpdate).Error
		if err != nil {
			return txErr
		}

		txErr = tx.Model(&groupToUpdate).Association("GroupMembers").Unscoped().Replace(groupToUpdate.GroupMembers)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return models.Group{}, err
	}

	returnGroup, err := repository.GetGroupById(groupId, true, true, true)
	if err != nil {
		return models.Group{}, err
	}

	return returnGroup, nil
}

func (repository GroupRepository) GetGroupById(id string,
	preloadGroupMembers bool,
	createMissingGroupSettings bool,
	createMissingGroupReceiptSettings bool,
) (models.Group, error) {
	db := repository.GetDB()
	var group models.Group

	query := db.Model(models.Group{}).Where("id = ?", id)
	if preloadGroupMembers {
		query = query.Preload("GroupMembers")
	}

	// TODO: Fix this repository call to take a preload string instead of a bool
	query = query.
		Preload("GroupSettings.SubjectLineRegexes").
		Preload("GroupSettings.EmailWhiteList").
		Preload("GroupSettings.SystemEmail").
		Preload("GroupSettings.Prompt").
		Preload("GroupSettings.FallbackPrompt").
		Preload("GroupReceiptSettings")

	err := query.First(&group).Error
	if err != nil {
		return models.Group{}, err
	}

	if group.GroupSettings.ID == 0 && createMissingGroupSettings {
		groupSettingsRepository := NewGroupSettingsRepository(repository.TX)

		groupSettings := models.GroupSettings{
			GroupId: group.ID,
		}

		_, err := groupSettingsRepository.CreateGroupSettings(groupSettings)
		if err != nil {
			return models.Group{}, err
		}
	}

	if group.GroupReceiptSettings.ID == 0 && createMissingGroupReceiptSettings {
		groupReceiptSettingsRepository := NewGroupReceiptSettingsRepository(repository.TX)

		_, err := groupReceiptSettingsRepository.CreateGroupReceiptSettings(group.ID)
		if err != nil {
			return models.Group{}, err
		}
	}

	return group, nil
}

func (repository GroupRepository) CreateAllGroup(userId uint) (models.Group, error) {
	group := commands.UpsertGroupCommand{
		Name:       "All",
		IsAllGroup: true,
	}

	allGroup, err := repository.CreateGroup(group, userId)
	if err != nil {
		return models.Group{}, err
	}

	return allGroup, nil
}

func (repository GroupRepository) IsAllGroup(groupId uint) (bool, error) {
	var group models.Group
	err := db.Where("id = ?", groupId).First(&group).Select("is_all_group").Error
	if err != nil {
		return false, err
	}

	return group.IsAllGroup, nil
}
