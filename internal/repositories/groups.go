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

	err = db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&group).Updates(&group).Error
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

	err := query.First(&group).Error
	if err != nil {
		return models.Group{}, err
	}

	return group, nil
}
