package repositories

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"

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

func (repository GroupRepository) CreateGroup(group models.Group, userId uint) (models.Group, error) {
	db := repository.GetDB()
	var returnGroup models.Group
	err := db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)
		groupSettingsRepository := NewGroupSettingsRepository(tx)

		txErr := tx.Model(&group).Create(&group).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupMember := models.GroupMember{
			UserID:    userId,
			GroupID:   group.ID,
			GroupRole: models.OWNER,
		}

		txErr = tx.Model(&groupMember).Create(&groupMember).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupSettings := models.GroupSettings{
			GroupId: group.ID,
		}

		_, txErr = groupSettingsRepository.CreateGroupSettings(groupSettings)
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

	err = repository.GetDB().Model(models.Group{}).Where("id = ?", group.ID).Preload("GroupMembers").Find(&returnGroup).Error
	if err != nil {
		return models.Group{}, err
	}

	return returnGroup, nil
}

func (repository GroupRepository) UpdateGroup(group models.Group, groupId string) (models.Group, error) {
	db := repository.GetDB()

	u64Id, err := simpleutils.StringToUint64(groupId)
	if err != nil {
		return models.Group{}, err
	}

	group.ID = uint(u64Id)

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&group).Omit("ID", "is_all_group").Updates(&group).Error
		if err != nil {
			return txErr
		}

		txErr = tx.Model(&group).Association("GroupMembers").Unscoped().Replace(group.GroupMembers)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return models.Group{}, err
	}

	returnGroup, err := repository.GetGroupById(groupId, true)
	if err != nil {
		return models.Group{}, err
	}

	return returnGroup, nil
}

func (repository GroupRepository) GetGroupById(id string, preloadGroupMembers bool) (models.Group, error) {
	db := repository.GetDB()
	var group models.Group

	query := db.Model(models.Group{}).Where("id = ?", id)
	if preloadGroupMembers {
		query = query.Preload("GroupMembers")
	}

	// TODO: Fix this repository call to take a preload string instead of a bool
	query = query.Preload("GroupSettings.SubjectLineRegexes").Preload("GroupSettings.EmailWhiteList").Preload("GroupSettings.SystemEmail")

	err := query.First(&group).Error
	if err != nil {
		return models.Group{}, err
	}

	if group.GroupSettings.ID == 0 {
		groupSettingsRepository := NewGroupSettingsRepository(db)

		groupSettings := models.GroupSettings{
			GroupId: group.ID,
		}

		_, err := groupSettingsRepository.CreateGroupSettings(groupSettings)
		if err != nil {
			return models.Group{}, err
		}
	}

	return group, nil
}

func (repository GroupRepository) CreateAllGroup(userId uint) (models.Group, error) {
	group := models.Group{
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
