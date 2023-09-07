package repositories

import (
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type GroupMemberRepository struct {
	BaseRepository
}

func NewGroupMemberRepository(tx *gorm.DB) GroupMemberRepository {
	repository := GroupMemberRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

// Gets groupMembers that the user has access to
func (repository GroupMemberRepository) GetGroupMembersByUserId(userId string) ([]models.GroupMember, error) {
	db := repository.GetDB()
	var groupMembers []models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ?", userId).Find(&groupMembers).Error
	if err != nil {
		return nil, err
	}

	return groupMembers, nil
}

// Gets group ids that the user has access to
func (repository GroupMemberRepository) GetGroupIdsByUserId(userId string) ([]uint, error) {
	groupMembers, err := repository.GetGroupMembersByUserId(userId)
	if err != nil {
		return nil, err
	}
	result := make([]uint, len(groupMembers))

	for i := 0; i < len(groupMembers); i++ {
		result[i] = groupMembers[i].GroupID
	}

	return result, nil
}

func (repository GroupMemberRepository) GetGroupMemberByUserIdAndGroupId(userId string, groupId string) (models.GroupMember, error) {
	db := repository.GetDB()
	var groupMember models.GroupMember

	err := db.Model(models.GroupMember{}).Where("user_id = ? AND group_id = ?", userId, groupId).First(&groupMember).Error
	if err != nil {
		return models.GroupMember{}, err
	}

	return groupMember, nil
}

func (repository GroupMemberRepository) GetsGroupMembersByGroupId(groupId string) ([]models.GroupMember, error) {
	db := repository.GetDB()
	var groupMembers []models.GroupMember

	err := db.Model(models.GroupMember{}).Where("group_id = ?", groupId).Find(&groupMembers).Error
	if err != nil {
		return nil, err
	}

	return groupMembers, nil
}
