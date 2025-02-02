package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
)

type GroupSettingsRepository struct {
	BaseRepository
}

func NewGroupSettingsRepository(tx *gorm.DB) GroupSettingsRepository {
	repository := GroupSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository GroupSettingsRepository) GetGroupSettingsById(id string) (models.GroupSettings, error) {
	var groupSettings models.GroupSettings
	err := repository.GetDB().
		Model(&models.GroupSettings{}).
		Preload(clause.Associations).
		Where("id = ?", id).
		First(&groupSettings).
		Error

	return groupSettings, err
}

func (repository GroupSettingsRepository) GetAllGroupSettings(queryWhere string, whereArgs ...any) ([]models.GroupSettings, error) {
	db := repository.GetDB()
	var groupSettings []models.GroupSettings

	query := db.Model(&models.GroupSettings{}).Preload(clause.Associations)
	if queryWhere != "" {
		query = query.Where(queryWhere, whereArgs...)
	}

	err := query.Find(&groupSettings).Error
	if err != nil {
		return nil, err
	}

	return groupSettings, nil
}

func (repository GroupSettingsRepository) CreateGroupSettings(groupSettings models.GroupSettings) (models.GroupSettings, error) {
	db := repository.GetDB()

	err := db.Model(&groupSettings).Create(&groupSettings).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	return groupSettings, nil
}

func (repository GroupSettingsRepository) UpdateGroupSettings(groupId string, command commands.UpdateGroupSettingsCommand) (models.GroupSettings, error) {
	db := repository.GetDB()

	var groupSettings models.GroupSettings

	err := db.Model(&models.GroupSettings{}).Where("group_id = ?", groupId).First(&groupSettings).Preload(clause.Associations).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	groupSettings.SystemEmailId = command.SystemEmailId
	groupSettings.EmailIntegrationEnabled = command.EmailIntegrationEnabled
	groupSettings.EmailDefaultReceiptStatus = command.EmailDefaultReceiptStatus
	groupSettings.EmailDefaultReceiptPaidById = command.EmailDefaultReceiptPaidById
	groupSettings.SubjectLineRegexes = command.SubjectLineRegexes
	groupSettings.EmailWhiteList = command.EmailWhiteList
	groupSettings.PromptId = command.PromptId
	groupSettings.FallbackPromptId = command.FallbackPromptId

	err = db.Transaction(func(tx *gorm.DB) error {

		err = tx.Session(&gorm.Session{FullSaveAssociations: true}).Select("*").Model(&groupSettings).Updates(&groupSettings).Error
		if err != nil {
			return err
		}

		err = tx.Model(&groupSettings).Association("SubjectLineRegexes").Replace(&groupSettings.SubjectLineRegexes)
		if err != nil {
			return err
		}

		err = tx.Model(&groupSettings).Association("EmailWhiteList").Replace(&groupSettings.EmailWhiteList)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.GroupSettings{}, err
	}

	return groupSettings, nil
}
