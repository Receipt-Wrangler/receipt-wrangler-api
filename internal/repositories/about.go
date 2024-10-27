package repositories

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/structs"
)

type AboutRepository struct {
	BaseRepository
}

func NewAboutRepository(tx *gorm.DB) AboutRepository {
	repository := AboutRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository AboutRepository) GetAboutData() (structs.About, error) {
	envVersion := "latest"
	envBuildDate := ""

	about := structs.About{
		Version:   envVersion,
		BuildDate: envBuildDate,
	}

	return about, nil
}
