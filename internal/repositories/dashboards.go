package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

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

	groupId, _ = utils.StringToUint(command.GroupId)

	for i, widget := range command.Widgets {
		widgets[i] = models.Widget{
			Name:          widget.Name,
			WidgetType:    widget.WidgetType,
			Configuration: widget.Configuration,
		}
	}

	dashboard := models.Dashboard{
		UserID:  userId,
		Name:    command.Name,
		GroupID: groupId,
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

func (repository *DashboardRepository) GetDashboardById(dashboardId uint) (models.Dashboard, error) {
	db := repository.GetDB()
	var dashboard models.Dashboard

	err := db.Model(models.Dashboard{}).Preload(clause.Associations).First(&dashboard, dashboardId).Error
	if err != nil {
		return models.Dashboard{}, err
	}

	return dashboard, nil
}

func (repository *DashboardRepository) DeleteDashboardById(dashboardId uint) error {
	db := repository.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		err := db.Delete(&models.Widget{}, "dashboard_id = ?", dashboardId).Error
		if err != nil {
			return err
		}

		err = db.Delete(models.Dashboard{}, dashboardId).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (repository *DashboardRepository) UpdateDashboardById(dashboardId uint, command commands.UpsertDashboardCommand) (models.Dashboard, error) {
	db := repository.GetDB()
	groupId, err := utils.StringToUint(command.GroupId)
	if err != nil {
		return models.Dashboard{}, err
	}

	db.Transaction(func(tx *gorm.DB) error {
		widgets := make([]models.Widget, len(command.Widgets))
		for i, widget := range command.Widgets {

			if widget.Configuration == nil {
				widget.Configuration = []byte("{}")
			}

			widgets[i] = models.Widget{
				DashboardId:   dashboardId,
				Name:          widget.Name,
				WidgetType:    widget.WidgetType,
				Configuration: widget.Configuration,
			}
		}

		dashboard := models.Dashboard{
			BaseModel: models.BaseModel{
				ID: dashboardId,
			},
			Name:    command.Name,
			GroupID: groupId,
		}

		if db.Model(&dashboard).Where("id = ?", dashboardId).Updates(&dashboard).Error != nil {
			return err
		}

		if db.Model(&dashboard).Association("Widgets").Unscoped().Replace(widgets) != nil {
			return err
		}

		return nil
	})

	updatedDashboard, err := repository.GetDashboardById(dashboardId)
	if err != nil {
		return models.Dashboard{}, err
	}

	return updatedDashboard, nil
}

func (repository *DashboardRepository) GetDashboardsByGroupId(groupId uint) ([]models.Dashboard, error) {
	db := repository.GetDB()
	var dashboards []models.Dashboard
	err := db.Where("group_id = ?", groupId).Find(&dashboards).Error
	if err != nil {
		return []models.Dashboard{}, err
	}

	return dashboards, nil
}
