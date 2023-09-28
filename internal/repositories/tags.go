package repositories

import (
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

func (repository TagsRepository) CreateTag(command commands.TagUpsertCommand) (models.Tag, error) {
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

	query := repository.Sort(db, pagedRequestCommand.OrderBy, pagedRequestCommand.SortDirection)
	query = query.Scopes(repository.Paginate(pagedRequestCommand.Page, pagedRequestCommand.PageSize))
	query = query.Table("receipt_tags").
		Select("*, COUNT(DISTINCT receipt_tags.receipt_id) as NumberOfReceipts").
		Joins("RIGHT JOIN tags ON receipt_tags.tag_id = tags.id").
		Group("receipt_tags.tag_id, tags.name")

	err := query.Scan(&tags).Error
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (repository TagsRepository) UpdateTag(tagId string, command commands.TagUpsertCommand) (models.Tag, error) {
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
