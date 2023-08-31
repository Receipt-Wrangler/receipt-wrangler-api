package repositories

import (
	"errors"
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReceiptRepository struct {
	BaseRepository
}

func NewReceiptRepository(tx *gorm.DB) ReceiptRepository {
	repository := ReceiptRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository ReceiptRepository) CreateReceipt(receipt models.Receipt, createdByUserID uint) (models.Receipt, error) {
	db := GetDB()
	notificationRepository := NewNotificationRepository(nil)

	receipt.CreatedBy = &createdByUserID

	err := db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)
		notificationRepository.SetTransaction(tx)
		err := tx.Model(models.Receipt{}).Select("*").Create(&receipt).Error
		if err != nil {
			return err
		}

		var userIdsToOmit []interface{} = make([]interface{}, 1)
		userIdsToOmit = append(userIdsToOmit, *receipt.CreatedBy)

		notificationBody := fmt.Sprintf("A receipt has been added in the group %s. Check it out! %s", BuildParamaterisedString("groupId", receipt.GroupId, "name", "string"), BuildParamaterisedString("receiptId", receipt.ID, "", "link"))
		notificationRepository.SendNotificationToGroup(receipt.GroupId, "Receipt Uploaded", notificationBody, models.NOTIFICATION_TYPE_NORMAL, userIdsToOmit)

		repository.ClearTransaction()
		notificationRepository.ClearTransaction()
		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (repository ReceiptRepository) GetReceiptById(receiptId string) (models.Receipt, error) {
	db := GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).First(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (repository ReceiptRepository) GetPagedReceiptsByGroupId(userId uint, groupId string, pagedRequest commands.ReceiptPagedRequestCommand) ([]models.Receipt, int64, error) {
	db := GetDB()
	var receipts []models.Receipt
	var count int64

	// Start query
	query := db.Table("receipts")

	// Filter receipts by group
	if groupId == "all" {
		groupMemberRepository := NewGroupMemberRepository(nil)
		groupIds, err := groupMemberRepository.GetGroupIdsByUserId(simpleutils.UintToString(userId))
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("group_id IN ?", groupIds)
	} else {
		query = query.Where("group_id = ?", groupId)
	}

	// Set order by
	if repository.isTrustedValue(pagedRequest) {
		orderBy := pagedRequest.OrderBy
		switch pagedRequest.OrderBy {

		case "paidBy":
			orderBy = "paid_by_user_id"
		case "resolvedDate":
			orderBy = "resolved_date"
		}
		query = query.Order(orderBy + " " + pagedRequest.SortDirection)
	} else {
		return nil, 0, errors.New("untrusted value " + pagedRequest.OrderBy + " " + pagedRequest.SortDirection)
	}

	// Apply filter

	// Name
	name := pagedRequest.Filter.Name.Value.(string)
	if len(name) > 0 {
		query = repository.buildFilterQuery(query, name, pagedRequest.Filter.Name.Operation, "name", false)
	}

	// Date
	date := pagedRequest.Filter.Date.Value.(string)
	if len(date) > 0 {
		query = repository.buildFilterQuery(query, date, pagedRequest.Filter.Date.Operation, "date", false)
	}

	// Paid By
	paidBy := pagedRequest.Filter.PaidBy.Value.([]interface{})
	if len(paidBy) > 0 {
		query = repository.buildFilterQuery(query, paidBy, pagedRequest.Filter.PaidBy.Operation, "paid_by_user_id", true)
	}

	// Categories
	categories := pagedRequest.Filter.Categories.Value.([]interface{})
	if len(categories) > 0 {
		if pagedRequest.Filter.Categories.Operation == commands.CONTAINS {
			query = query.Where("id IN (?)", db.Table("receipt_categories").Select("receipt_id").Where("category_id IN ?", categories))
		}
	}

	// Tags
	tags := pagedRequest.Filter.Tags.Value.([]interface{})
	if len(tags) > 0 {
		if pagedRequest.Filter.Tags.Operation == commands.CONTAINS {
			query = query.Where("id IN (?)", db.Table("receipt_tags").Select("receipt_id").Where("tag_id IN ?", tags))
		}
	}

	// Amount
	amount := pagedRequest.Filter.Amount.Value.(float64)
	if amount > 0 {
		query = repository.buildFilterQuery(query, amount, pagedRequest.Filter.Amount.Operation, "amount", false)
	}

	// Status
	status := pagedRequest.Filter.Status.Value.([]interface{})
	if len(status) > 0 {
		query = repository.buildFilterQuery(query, status, pagedRequest.Filter.Status.Operation, "status", true)
	}

	// Resolved Date
	resolvedDate := pagedRequest.Filter.ResolvedDate.Value.(string)
	if len(resolvedDate) > 0 {
		query = repository.buildFilterQuery(query, resolvedDate, pagedRequest.Filter.ResolvedDate.Operation, "resolved_date", false)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	query = query.Scopes(repository.Paginate(pagedRequest.Page, pagedRequest.PageSize)).Preload("Tags").Preload("Categories")

	// Run Query
	err = query.Find(&receipts).Error
	if err != nil {
		return nil, 0, err
	}

	return receipts, count, nil
}

func (repository ReceiptRepository) buildFilterQuery(runningQuery *gorm.DB, value interface{}, operation commands.FilterOperation, fieldName string, isArray bool) *gorm.DB {

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

	return runningQuery
}

func (repository ReceiptRepository) isTrustedValue(pagedRequest commands.ReceiptPagedRequestCommand) bool {
	orderByTrusted := []interface{}{"date", "name", "paidBy", "amount", "categories", "tags", "status", "resolvedDate"}
	directionTrusted := []interface{}{"asc", "desc", ""}

	isOrderByTrusted := utils.Contains(orderByTrusted, pagedRequest.OrderBy)
	isDirectionTrusted := utils.Contains(directionTrusted, pagedRequest.SortDirection)

	return isOrderByTrusted && isDirectionTrusted
}

func (repository ReceiptRepository) GetReceiptGroupIdByReceiptId(id string) (uint, error) {
	db := GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("group_id").Find(&receipt).Error
	if err != nil {
		return 0, err
	}

	return receipt.GroupId, nil
}

func (repository ReceiptRepository) GetFullyLoadedReceiptById(id string) (models.Receipt, error) {
	db := GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (repository ReceiptRepository) GetReceiptsByGroupIds(groupIds []string, querySelect string, queryPreload string) ([]models.Receipt, error) {
	db := GetDB()
	var receipts []models.Receipt

	query := db.Model(models.Receipt{}).Where("group_id IN ?", groupIds).Select(querySelect)
	if len(queryPreload) > 0 {
		query = query.Preload(queryPreload)
	}

	err := query.Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	return receipts, nil
}
