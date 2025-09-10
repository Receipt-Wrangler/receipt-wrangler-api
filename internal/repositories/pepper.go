package repositories

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/models"
)

type PepperRepository struct {
	BaseRepository
}

func NewPepperRepository(tx *gorm.DB) PepperRepository {
	repository := CommentRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository *PepperRepository) CreatePepper(pepper models.Pepper) error {
	return db.Model(&pepper).Create(pepper).Error
}
