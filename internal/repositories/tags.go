package repositories

import (
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

type TagsRepository struct {
	BaseRepository
}

func NewTagsRepository(tx *gorm.DB) TagsRepository {
	repository := TagsRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository TagsRepository) GetAllTags(querySelect string) ([]models.Tag, error) {
	db := repository.GetDB()
	var tags []models.Tag

	err := db.Table("tags").Select(querySelect).Find(&tags).Error
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (repository TagsRepository) CreateTag(command commands.UpsertTagCommand) (models.Tag, error) {
	db := repository.GetDB()
	tag := models.Tag{}

	tag.Name = command.Name
	tag.Description = command.Description

	err := db.Model(&tag).Create(&tag).Error
	if err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func (repository TagsRepository) GetAllPagedTags(pagedRequestCommand commands.PagedRequestCommand) ([]models.TagView, error) {
	db := repository.GetDB()
	var tags []models.TagView
	quotedAlias := "\"NumberOfReceipts\""

	if pagedRequestCommand.OrderBy == "numberOfReceipts" {
		pagedRequestCommand.OrderBy = quotedAlias
	}

	query := repository.Sort(db, pagedRequestCommand.OrderBy, pagedRequestCommand.SortDirection)
	query = query.Scopes(repository.Paginate(pagedRequestCommand.Page, pagedRequestCommand.PageSize))
	selectString := fmt.Sprintf("tags.id, tags.name, tags.description, COUNT(DISTINCT receipt_tags.receipt_id) as %s", quotedAlias)
	query = query.Table("tags").
		Select(selectString).
		Joins("LEFT JOIN receipt_tags ON tags.id = receipt_tags.tag_id").
		Group("tags.id, tags.name")

	err := query.Scan(&tags).Error
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (repository TagsRepository) UpdateTag(tagId string, command commands.UpsertTagCommand) (models.Tag, error) {
	db := repository.GetDB()
	var updatedTag models.Tag

	err := db.Model(models.Tag{}).Where("id = ?", tagId).Updates(command).Error
	if err != nil {
		return models.Tag{}, err
	}

	err = db.Model(models.Tag{}).Where("id = ?", tagId).Find(&updatedTag).Error
	if err != nil {
		return models.Tag{}, err
	}

	return updatedTag, nil
}

func (repository TagsRepository) DeleteTag(tagId uint) error {
	db := repository.GetDB()

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&models.ReceiptTag{}, "tag_id = ?", tagId).Error
		if err != nil {
			return err
		}

		err = tx.Where("id = ?", tagId).Delete(&models.Tag{}).Error
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
