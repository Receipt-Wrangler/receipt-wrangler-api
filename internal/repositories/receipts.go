package repositories

import (
	"errors"
	"fmt"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PaginateReceipts(pagedRequest structs.PagedRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page := pagedRequest.Page
		if page <= 0 {
			page = 1
		}

		pageSize := pagedRequest.PageSize
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

func GetReceiptById(receiptId string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func GetPagedReceiptsByGroupId(userId uint, groupId string, pagedRequest structs.PagedRequest) ([]models.Receipt, error) {
	db := db.GetDB()
	var receipts []models.Receipt

	// Start query
	query := db.Scopes(PaginateReceipts(pagedRequest)).Model(models.Receipt{}).Preload("Tags").Preload("Categories")

	// Filter receipts by group
	if groupId == "all" {
		groupIds, err := GetGroupIdsByUserId(simpleutils.UintToString(userId))
		if err != nil {
			return nil, err
		}
		query = query.Where("group_id IN ?", groupIds)
	} else {
		query = query.Where("group_id = ?", groupId)
	}

	// Set order by
	if isTrustedValue(pagedRequest) {
		orderBy := pagedRequest.OrderBy
		switch pagedRequest.OrderBy {

		case "paidBy":
			orderBy = "paid_by_user_id"
		case "resolvedDate":
			orderBy = "resolved_date"
		}
		query = query.Order(orderBy + " " + pagedRequest.SortDirection)
	} else {
		return nil, errors.New("untrusted value " + pagedRequest.OrderBy + " " + pagedRequest.SortDirection)
	}

	// Apply filter

	// Name
	name := pagedRequest.Filter.Name.Value.(string)
	if len(name) > 0 {
		query = buildFilterQuery(query, name, pagedRequest.Filter.Name.Operation, "name", false)
	}

	// Date
	date := pagedRequest.Filter.Date.Value.(string)
	if len(date) > 0 {
		query = buildFilterQuery(query, date, pagedRequest.Filter.Date.Operation, "date", false)
	}

	// Paid By
	paidBy := pagedRequest.Filter.PaidBy.Value.([]interface{})
	if len(paidBy) > 0 {
		query = buildFilterQuery(query, paidBy, pagedRequest.Filter.PaidBy.Operation, "paid_by_user_id", true)
	}

	// Amount
	amount := pagedRequest.Filter.Amount.Value.(float64)
	if amount > 0 {
		query = buildFilterQuery(query, amount, pagedRequest.Filter.Amount.Operation, "amount", false)
	}

	status := pagedRequest.Filter.Status.Value.([]interface{})
	if len(status) > 0 {
		operation := pagedRequest.Filter.Status.Operation

		if operation == structs.CONTAINS {
			query = query.Where("status IN ?", status)
		}
	}

	// Run Query
	err := query.Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	return receipts, nil
}

func buildFilterQuery(runningQuery *gorm.DB, value interface{}, operation structs.FilterOperation, fieldName string, isArray bool) *gorm.DB {

	if operation == structs.EQUALS && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v = ?", fieldName), value)
	}

	if operation == structs.CONTAINS && !isArray {
		searchValue := value.(string)
		searchValue = "%" + searchValue + "%"
		return runningQuery.Where(fmt.Sprintf("%v LIKE ?", fieldName), searchValue)
	}

	if operation == structs.CONTAINS && isArray {
		return runningQuery.Where(fmt.Sprintf("%v IN ?", fieldName), value)
	}

	if operation == structs.GREATER_THAN && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v > ?", fieldName), value)
	}

	if operation == structs.LESS_THAN && !isArray {
		return runningQuery.Where(fmt.Sprintf("%v < ?", fieldName), value)
	}

	return runningQuery
}

func isTrustedValue(pagedRequest structs.PagedRequest) bool {
	orderByTrusted := []interface{}{"date", "name", "paidBy", "amount", "categories", "tags", "status", "resolvedDate"}
	directionTrusted := []interface{}{"asc", "desc", ""}

	isOrderByTrusted := utils.Contains(orderByTrusted, pagedRequest.OrderBy)
	isDirectionTrusted := utils.Contains(directionTrusted, pagedRequest.SortDirection)

	return isOrderByTrusted && isDirectionTrusted
}

func GetGroupReceiptCount(userId uint, groupId string) (int64, error) {
	db := db.GetDB()
	var result int64
	var err error

	if groupId == "all" {
		groupIds, err := GetGroupIdsByUserId(simpleutils.UintToString(userId))
		if err != nil {
			return 0, err
		}

		err = db.Model(models.Receipt{}).Where("group_id IN ?", groupIds).Count(&result).Error
	} else {
		err = db.Model(models.Receipt{}).Where("group_id = ?", groupId).Count(&result).Error
	}

	if err != nil {
		return 0, err
	}

	return result, nil
}

func GetReceiptGroupIdByReceiptId(id string) (uint, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("group_id").Find(&receipt).Error
	if err != nil {
		return 0, err
	}

	return receipt.GroupId, nil
}

func GetFullyLoadedReceiptById(id string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func GetReceiptsByGroupIds(groupIds []string) ([]models.Receipt, error) {
	db := db.GetDB()
	var receipts []models.Receipt

	err := db.Model(models.Receipt{}).Where("group_id IN ?", groupIds).Preload(clause.Associations).Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	return receipts, nil
}
