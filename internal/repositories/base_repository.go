package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
)

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

func (repository *BaseRepository) SetTransaction(tx *gorm.DB) {
	repository.TX = tx
}

func (repository *BaseRepository) ClearTransaction() {
	repository.TX = nil
}

func (repository BaseRepository) Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pageSize == -1 {
			return db
		}

		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func (repository BaseRepository) Sort(db *gorm.DB, orderBy string, sortDirection commands.SortDirection) *gorm.DB {
	desc := false
	if sortDirection == commands.DESCENDING {
		desc = true
	}

	return db.Order(clause.OrderByColumn{
		Column:  clause.Column{Name: orderBy},
		Desc:    desc,
		Reorder: false,
	})
}

func (repository BaseRepository) GetCount(table string, queryWhere string) (int64, error) {
	db := repository.GetDB()
	var result int64
	err := db.Table(table).Where(queryWhere).Count(&result).Error

	return result, err
}
