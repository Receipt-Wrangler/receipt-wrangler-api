package repositories

import (
	"gorm.io/gorm"
	"os"
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
	envVersion := os.Getenv("VERSION")
	envBuildDate := os.Getenv("BUILD_DATE")

	if len(envVersion) == 0 {
		envVersion = "unknown"
	}

	if len(envBuildDate) == 0 {
		envBuildDate = "unknown"
	}

	about := structs.About{
		Version:   envVersion,
		BuildDate: envBuildDate,
	}

	return about, nil
}
