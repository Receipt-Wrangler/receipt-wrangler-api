package repositories

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type SystemSettingsRepository struct {
	BaseRepository
}

func NewSystemSettingsRepository(tx *gorm.DB) SystemSettingsRepository {
	repository := SystemSettingsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository SystemSettingsRepository) GetSystemSettings() (models.SystemSettings, error) {
	db := repository.GetDB()
	var systemSettings models.SystemSettings
	var count int64

	err := db.Model(&models.SystemSettings{}).Count(&count).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	if count == 0 {
		err = db.Model(&models.SystemSettings{}).Create(&models.SystemSettings{
			BaseModel: models.BaseModel{
				ID: 1,
			},
		}).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	err = db.Model(&models.SystemSettings{}).Preload(clause.Associations).Preload("TaskQueueConfigurations").First(&systemSettings).Error
	if err != nil {
		return models.SystemSettings{}, err
	}

	// NOTE: Eventually this can get deleted. This is to fix associations not working if ID Is 0
	if systemSettings.ID == 0 {
		err = db.Model(models.SystemSettings{}).Where("id = 0").Update("id", 1).Error
		if err != nil {
			return models.SystemSettings{}, err
		}
	}

	systemSettings.TaskQueueConfigurations = withMissingQueueConfigurations(systemSettings.TaskQueueConfigurations)

	return systemSettings, nil
}

// withMissingQueueConfigurations returns the persisted queue configurations with a
// default filled in for every queue name the list does not carry, ordered by
// models.GetQueueNames(). A persisted configuration always wins; only absent names
// are filled.
//
// An install that upgraded across a release which added a queue has rows for only
// the queues that existed when it last saved, and nothing backfills the rest: this
// read used to substitute defaults only when the list was EMPTY, and
// UpdateSystemSettings only ever UPDATEs by name. The settings form builds its rows
// from this response, so it submitted a short list and
// UpsertSystemSettingsCommand.Validate — which requires a configuration for every
// queue name — rejected the whole save with a 400, every field in it along with the
// queue list. The task server never surfaced the gap, because asynq_server.go falls
// back to the default priority for a missing configuration.
//
// A persisted row naming a queue this build does not know is dropped rather than
// carried through: QueueName.Value() would reject it on the way back down, so
// returning it could only make an otherwise valid save fail.
func withMissingQueueConfigurations(persisted []models.TaskQueueConfiguration) []models.TaskQueueConfiguration {
	persistedByName := make(map[models.QueueName]models.TaskQueueConfiguration, len(persisted))
	for _, configuration := range persisted {
		persistedByName[configuration.Name] = configuration
	}

	defaultsByName := models.GetDefaultQueueConfigurationMap()
	queueNames := models.GetQueueNames()
	configurations := make([]models.TaskQueueConfiguration, 0, len(queueNames))

	for _, queueName := range queueNames {
		if configuration, ok := persistedByName[queueName]; ok {
			configurations = append(configurations, configuration)
			continue
		}

		configurations = append(configurations, defaultsByName[queueName])
	}

	return configurations
}

func (repository SystemSettingsRepository) GetSystemReceiptProcessingSettings() (structs.SystemReceiptProcessingSettings, error) {
	systemSettings, err := repository.GetSystemSettings()
	if err != nil {
		return structs.SystemReceiptProcessingSettings{}, err
	}

	if systemSettings.ReceiptProcessingSettings.ID == 0 {
		return structs.SystemReceiptProcessingSettings{}, errors.New("ReceiptProcessingSettings do not exist")
	}

	return structs.SystemReceiptProcessingSettings{
		ReceiptProcessingSettings:         systemSettings.ReceiptProcessingSettings,
		FallbackReceiptProcessingSettings: systemSettings.FallbackReceiptProcessingSettings,
	}, nil
}

func (repository SystemSettingsRepository) UpdateSystemSettings(command commands.UpsertSystemSettingsCommand) (models.SystemSettings, error) {
	db := repository.GetDB()

	existingSettings, err := repository.GetSystemSettings()
	if err != nil {
		return models.SystemSettings{}, err
	}

	updatedSettings, err := command.ToSystemSettings(existingSettings.ID)
	if err != nil {
		return models.SystemSettings{}, err
	}

	// Select("*") below writes every column, so any field the request omitted
	// would be persisted as its zero value. Two things guard the lifetimes:
	// the columns the caller did not send are dropped from the UPDATE entirely
	// (so a concurrent update that DID set one is never clobbered by a value we
	// read before the write), and the in-memory copy is back-filled so the
	// response echoes the stored value rather than a 0.
	omittedColumns := append([]string{"TaskQueueConfigurations"}, command.OmittedColumns()...)
	command.ApplyOmittedValues(existingSettings, &updatedSettings)

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Model(&updatedSettings).Select("*").Omit(omittedColumns...).Where("id = ?", existingSettings.ID).Updates(&updatedSettings).Error
		if txErr != nil {
			return txErr
		}

		var configCount int64
		txErr = tx.Model(&models.TaskQueueConfiguration{}).Count(&configCount).Error
		if txErr != nil {
			return txErr
		}

		if configCount == 0 {
			txErr = tx.Model(&updatedSettings).
				Where("id = ?", existingSettings.ID).
				Association("TaskQueueConfigurations").
				Replace(&updatedSettings.TaskQueueConfigurations)
			if txErr != nil {
				return txErr
			}
		} else {
			var persistedNames []string
			txErr = tx.Model(&models.TaskQueueConfiguration{}).Pluck("name", &persistedNames).Error
			if txErr != nil {
				return txErr
			}

			// Which names already have a row is read up front rather than inferred
			// from RowsAffected, which a no-change UPDATE reports as 0 on MySQL.
			existingNames := make(map[models.QueueName]bool, len(persistedNames))
			for _, name := range persistedNames {
				existingNames[models.QueueName(name)] = true
			}

			for _, config := range updatedSettings.TaskQueueConfigurations {
				// A queue added by a release this install upgraded across has no row
				// yet, and GetSystemSettings has just filled its default in for the
				// caller. Without the insert the UPDATE matches nothing, the submitted
				// priority is silently discarded, and the missing row never heals.
				if !existingNames[config.Name] {
					newConfiguration := models.TaskQueueConfiguration{
						Name:             config.Name,
						Priority:         config.Priority,
						SystemSettingsId: existingSettings.ID,
					}

					txErr = tx.Create(&newConfiguration).Error
					if txErr != nil {
						return txErr
					}

					continue
				}

				txErr = tx.Model(&models.TaskQueueConfiguration{}).Where("name = ?", config.Name).Updates(&models.TaskQueueConfiguration{
					Priority: config.Priority,
				}).Error

				if txErr != nil {
					return txErr
				}
			}
		}

		return nil
	})

	if err != nil {
		// The transaction rolled back, so nothing was saved. Swallowing this
		// returned a 200 with an empty body and lost the change without a
		// signal -- and it silently disarmed every caller's error check.
		return models.SystemSettings{}, err
	}

	return updatedSettings, nil
}
