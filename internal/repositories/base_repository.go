package repositories

import "gorm.io/gorm"

type BaseRepository struct {
	DB *gorm.DB
	TX *gorm.DB
}

func (repository BaseRepository) GetDB() *gorm.DB {
	if repository.TX != nil {
		return repository.TX
	}

	return repository.DB
}
