package services

import "gorm.io/gorm"

type BaseService struct {
	DB *gorm.DB
	TX *gorm.DB
}

func (repository BaseService) GetDB() *gorm.DB {
	if repository.TX != nil {
		return repository.TX
	}

	return repository.DB
}

func (repository *BaseService) SetTransaction(tx *gorm.DB) {
	repository.TX = tx
}

func (repository *BaseService) ClearTransaction() {
	repository.TX = nil
}
