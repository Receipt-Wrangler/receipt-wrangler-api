package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (repository GroupSettingsRepository) GetAllGroupSettings(queryWhere string, whereArgs any) ([]models.GroupSettings, error) {
	db := repository.GetDB()
	var groupSettings []models.GroupSettings

	query := db.Model(&models.GroupSettings{}).Preload(clause.Associations)
	if queryWhere != "" {
		query = query.Where(queryWhere, whereArgs)
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

	err := db.Model(&models.GroupSettings{}).Where("group_id = ?", groupId).First(&groupSettings).Error
	if err != nil {
		return models.GroupSettings{}, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {

		err = tx.Model(&groupSettings).Select("*").Updates(map[string]interface{}{"email_to_read": command.EmailToRead, "email_integration_enabled": command.EmailIntegrationEnabled, "email_default_receipt_paid_by_id": command.EmailDefaultReceiptPaidById, "email_default_receipt_status": command.EmailDefaultReceiptStatus}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&groupSettings).Association("SubjectLineRegexes").Replace(&command.SubjectLineRegexes)
		if err != nil {
			return err
		}

		err = tx.Model(&groupSettings).Association("EmailWhiteList").Replace(&command.EmailWhiteList)
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
