package repositories

import (
	"fmt"
	"time"

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

func (repository BaseRepository) GetCount(table string, queryWhere string, args ...interface{}) (int64, error) {
	db := repository.GetDB()
	var result int64
	err := db.Table(table).Where(queryWhere, args...).Count(&result).Error

	return result, err
}

// BuildFilterQuery translates one {operation, value} filter field into a WHERE
// clause on fieldName. Shared by the receipt filter and the system task filter
// so the two cannot drift on what an operation means.
//
// fieldName is interpolated into the clause, so it MUST be a hardcoded column
// literal supplied by the caller — never a value taken from a request.
func (repository BaseRepository) BuildFilterQuery(runningQuery *gorm.DB, value interface{}, operation commands.FilterOperation, fieldName string, isArray bool) *gorm.DB {
	if operation == commands.EQUALS && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v = ?", fieldName), value)
	}

	if operation == commands.CONTAINS && !isArray {
		searchValue := value.(string)
		searchValue = "%" + searchValue + "%"
		return runningQuery.Where(fmt.Sprintf("%v LIKE ?", fieldName), searchValue)
	}

	if operation == commands.CONTAINS && isArray {
		return runningQuery.Where(fmt.Sprintf("%v IN ?", fieldName), value)
	}

	if operation == commands.GREATER_THAN && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v > ?", fieldName), value)
	}

	if operation == commands.LESS_THAN && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v < ?", fieldName), value)
	}

	if operation == commands.BETWEEN {
		arrayValue, ok := value.([]interface{})
		if !ok || len(arrayValue) != 2 {
			return runningQuery
		}

		return runningQuery.Where(fmt.Sprintf("%v >= ? AND %v <= ?", fieldName, fieldName), arrayValue[0], arrayValue[1])
	}

	if operation == commands.WITHIN_CURRENT_MONTH {
		now := time.Now()
		beginningOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

		return runningQuery.Where(fmt.Sprintf("%v >= ? AND %v <= ?", fieldName, fieldName), beginningOfMonth, endOfToday)
	}

	return runningQuery
}
