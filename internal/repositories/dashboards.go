package repositories

import "gorm.io/gorm"

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
