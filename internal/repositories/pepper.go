package repositories

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
)

type PepperRepository struct {
	BaseRepository
}

func NewPepperRepository(tx *gorm.DB) PepperRepository {
	repository := PepperRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository *PepperRepository) CreatePepper(pepper models.Pepper) error {
	return repository.GetDB().Model(&pepper).Create(&pepper).Error
}

func (repository *PepperRepository) GetPepper() (*models.Pepper, error) {
	var pepper models.Pepper

	var err = repository.GetDB().Model(&pepper).First(&pepper).Error
	if err != nil {
		return nil, err
	}

	return &pepper, nil
}
