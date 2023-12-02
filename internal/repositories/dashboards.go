package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DashboardRepository struct {
	BaseRepository
}

func NewDashboardRepository(tx *gorm.DB) DashboardRepository {
	repository := DashboardRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository *DashboardRepository) CreateDashboard(command commands.UpsertDashboardCommand, userId uint) (models.Dashboard, error) {
	db := repository.GetDB()
	widgets := make([]models.Widget, len(command.Widgets))
	var groupId uint

	groupId, _ = simpleutils.StringToUint(command.GroupId)

	for i, widget := range command.Widgets {
		configuration := []byte("{}")

		widgets[i] = models.Widget{
			Name:          widget.Name,
			WidgetType:    widget.WidgetType,
			Configuration: configuration,
		}
	}

	dashboard := models.Dashboard{
		UserID:  userId,
		Name:    command.Name,
		GroupID: &groupId,
		Widgets: widgets,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		err := db.Create(&dashboard).Error
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		return nil
	})
	if err != nil {
		return models.Dashboard{}, err
	}

	return dashboard, nil
}

func (repository *DashboardRepository) GetDashboardsForUserByGroup(userId uint, groupId uint) ([]models.Dashboard, error) {
	db := repository.GetDB()
	var dashboards []models.Dashboard

	err := db.Where("user_id = ? and group_id = ?", userId, groupId).Preload(clause.Associations).Find(&dashboards).Error
	if err != nil {
		return []models.Dashboard{}, err
	}

	return dashboards, nil
}
