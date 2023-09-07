package repositories

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

		txErr := repository.GetDB().Model(&group).Create(&group).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupMember := models.GroupMember{
			UserID:    userId,
			GroupID:   group.ID,
			GroupRole: models.OWNER,
		}

		txErr = repository.GetDB().Model(&groupMember).Create(&groupMember).Error
		if txErr != nil {
			repository.ClearTransaction()
			return txErr
		}

		groupSettings := models.GroupSettings{
			GroupId: group.ID,
		}

		txErr = repository.GetDB().Model(&groupSettings).Create(&groupSettings).Error
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
		txErr := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&group).Omit("ID").Updates(&group).Error
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
	query.Preload(clause.Associations)

	err := query.First(&group).Error
	if err != nil {
		return models.Group{}, err
	}

	return group, nil
}
