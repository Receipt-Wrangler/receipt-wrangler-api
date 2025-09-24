package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"time"

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

func (repository ReceiptRepository) BeforeUpdateReceipt(currentReceipt models.Receipt, updatedReceipt models.Receipt) (err error) {
	db := repository.GetDB()
	if updatedReceipt.GroupId > 0 && currentReceipt.GroupId != updatedReceipt.GroupId && len(currentReceipt.ImageFiles) > 0 {
		var oldGroup models.Group
		var newGroup models.Group

		err = db.Table("groups").Where("id = ?", currentReceipt.GroupId).Select("id", "name").Find(&oldGroup).Error
		if err != nil {
			return err
		}

		err = db.Table("groups").Where("id = ?", updatedReceipt.GroupId).Select("id", "name").Find(&newGroup).Error
		if err != nil {
			return err
		}

		oldGroupPath, err := utils.BuildGroupPathString(utils.UintToString(oldGroup.ID), oldGroup.Name)
		if err != nil {
			return err
		}

		newGroupPath, err := utils.BuildGroupPathString(utils.UintToString(newGroup.ID), newGroup.Name)
		if err != nil {
			return err
		}

		for _, fileData := range currentReceipt.ImageFiles {
			filename := utils.BuildFileName(utils.UintToString(currentReceipt.ID), utils.UintToString(fileData.ID), fileData.Name)

			oldFilePath := filepath.Join(oldGroupPath, filename)
			newFilePathPath := filepath.Join(newGroupPath, filename)

			err := os.Rename(oldFilePath, newFilePathPath)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func createFailedUpdateSystemTask(command commands.UpsertSystemTaskCommand, err error) {
	endedAt := time.Now()
	command.EndedAt = &endedAt
	command.Status = models.SYSTEM_TASK_FAILED
	command.ResultDescription = err.Error()

	repository := NewSystemTaskRepository(nil)
	repository.CreateSystemTask(command)
}

func (repository ReceiptRepository) UpdateReceipt(id string, command commands.UpsertReceiptCommand, userId uint) (models.Receipt, error) {
	db := repository.GetDB()

	systemTaskResultDescription := map[string]interface{}{}
	var endedAt time.Time
	stringId, err := utils.StringToUint(id)
	if err != nil {
		return models.Receipt{}, err
	}
	var currentReceipt models.Receipt
	var ranByUserId = userId

	systemTask := commands.UpsertSystemTaskCommand{
		Type:                 models.RECEIPT_UPDATED,
		AssociatedEntityType: models.RECEIPT,
		AssociatedEntityId:   stringId,
		StartedAt:            time.Now(),
		EndedAt:              &endedAt,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		RanByUserId:          &ranByUserId,
	}

	updatedReceipt, err := command.ToReceipt()
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	err = db.Table("receipts").Where("id = ?", id).Preload(clause.Associations).Find(&currentReceipt).Error
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	systemTask.GroupId = &currentReceipt.GroupId
	systemTask.ReceiptId = &currentReceipt.ID

	// NOTE: ID and field used for afterReceiptUpdated
	updatedReceipt.ID = currentReceipt.ID
	updatedReceipt.ResolvedDate = currentReceipt.ResolvedDate
	before, err := currentReceipt.ToString()
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}
	systemTaskResultDescription["before"] = before

	err = db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)

		txErr := repository.BeforeUpdateReceipt(currentReceipt, updatedReceipt)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&currentReceipt).Updates(&updatedReceipt).Error
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("Tags").Replace(&updatedReceipt.Tags)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("Categories").Replace(&updatedReceipt.Categories)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("ReceiptItems").Replace(&updatedReceipt.ReceiptItems)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&currentReceipt).Association("CustomFields").Replace(&updatedReceipt.CustomFields)
		if txErr != nil {
			return txErr
		}

		for _, item := range updatedReceipt.ReceiptItems {
			txErr = tx.Model(&item).Association("Categories").Replace(&item.Categories)
			if txErr != nil {
				return txErr
			}

			txErr = tx.Model(&item).Association("Tags").Replace(&item.Tags)
			if txErr != nil {
				return txErr
			}

			txErr = tx.Model(&item).Association("LinkedItems").Replace(&item.LinkedItems)
			if txErr != nil {
				return txErr
			}

			// Update categories and tags for linked items
			for _, linkedItem := range item.LinkedItems {
				txErr = tx.Model(&linkedItem).Association("Categories").Replace(&linkedItem.Categories)
				if txErr != nil {
					return txErr
				}

				txErr = tx.Model(&linkedItem).Association("Tags").Replace(&linkedItem.Tags)
				if txErr != nil {
					return txErr
				}
			}
		}

		err = repository.AfterReceiptUpdated(&updatedReceipt)
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		return nil
	})
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	fullyLoadedReceipt, err := repository.GetFullyLoadedReceiptById(id)
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	after, err := fullyLoadedReceipt.ToString()
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	systemTaskResultDescription["after"] = after
	endedAt = time.Now()
	systemTask.EndedAt = &endedAt

	resultDescriptionBytes, err := json.Marshal(systemTaskResultDescription)
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}
	systemTask.ResultDescription = string(resultDescriptionBytes)

	systemTaskRepository := NewSystemTaskRepository(nil)
	_, err = systemTaskRepository.CreateSystemTask(systemTask)
	if err != nil {
		createFailedUpdateSystemTask(systemTask, err)
		return models.Receipt{}, err
	}

	return fullyLoadedReceipt, nil
}

// TODO: Delete categories/tags here associated with items before deleting the items mkay
func (repository ReceiptRepository) AfterReceiptUpdated(updatedReceipt *models.Receipt) error {
	db := repository.GetDB()

	// TODO: Move this  to a scheduled job
	// Clean up junction tables for orphaned items
	orphanedItemsSubquery := db.Table("items").Select("id").Where("receipt_id IS NULL")

	// Clean up item_linked_items junction table - remove associations where either side is orphaned
	err := db.Table("item_linked_items").Where("item_id IN (?) OR linked_item_id IN (?)",
		orphanedItemsSubquery,
		orphanedItemsSubquery,
	).Delete(&struct{}{}).Error
	if err != nil {
		return err
	}

	// Clean up item_categories junction table
	err = db.Table("item_categories").Where("item_id IN (?)",
		orphanedItemsSubquery,
	).Delete(&struct{}{}).Error
	if err != nil {
		return err
	}

	// Clean up item_tags junction table
	err = db.Table("item_tags").Where("item_id IN (?)",
		orphanedItemsSubquery,
	).Delete(&struct{}{}).Error
	if err != nil {
		return err
	}

	// TODO: Move this  to a scheduled job
	// Delete the orphaned items themselves
	err = db.Where("receipt_id IS NULL").Delete(&models.Item{}).Error
	if err != nil {
		return err
	}

	if updatedReceipt.ID > 0 && updatedReceipt.Status == models.RESOLVED && updatedReceipt.ResolvedDate == nil {
		now := time.Now().UTC()
		err = db.Table("receipts").Where("id = ?", updatedReceipt.ID).Update("resolved_date", now).Error
	} else if updatedReceipt.ID > 0 && updatedReceipt.Status != models.RESOLVED && updatedReceipt.ResolvedDate != nil {
		err = db.Table("receipts").Where("id = ?", updatedReceipt.ID).Update("resolved_date", nil).Error
	}
	if err != nil {
		return err
	}

	if updatedReceipt.Status == models.RESOLVED && updatedReceipt.ID > 0 {
		err := repository.UpdateItemsToStatus(updatedReceipt, models.ITEM_RESOLVED)
		if err != nil {
			return err
		}
	}

	if updatedReceipt.Status == models.DRAFT && updatedReceipt.ID > 0 {
		err := repository.UpdateItemsToStatus(updatedReceipt, models.ITEM_DRAFT)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repository ReceiptRepository) UpdateItemsToStatus(receipt *models.Receipt, status models.ItemStatus) error {
	db := repository.GetDB()
	var items []models.Item
	var itemIdsToUpdate []uint

	err := db.Table("items").Where("receipt_id = ?", receipt.ID).Find(&items).Error
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Status != status {
			itemIdsToUpdate = append(itemIdsToUpdate, item.ID)
		}
	}

	if len(itemIdsToUpdate) > 0 {
		err := db.Table("items").Where("id IN ?", itemIdsToUpdate).UpdateColumn("status", status).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (repository ReceiptRepository) CreateReceipt(
	command commands.UpsertReceiptCommand,
	createdByUserID uint,
	createSystemTask bool,
) (models.Receipt, error) {
	db := repository.GetDB()
	notificationRepository := NewNotificationRepository(nil)
	receipt, err := command.ToReceipt()
	if err != nil {
		return models.Receipt{}, err
	}

	if receipt.GroupId > 0 {
		receipt.CreatedBy = &createdByUserID
	}

	systemTask := commands.UpsertSystemTaskCommand{
		Type:                 models.RECEIPT_UPLOADED,
		AssociatedEntityType: models.RECEIPT,
		StartedAt:            time.Now(),
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		RanByUserId:          &createdByUserID,
	}

	// Extract linked items before creating receipt
	type LinkedItemData struct {
		ParentItemIndex int
		LinkedItem      models.Item
	}
	var linkedItemsData []LinkedItemData

	for i := range receipt.ReceiptItems {
		if len(receipt.ReceiptItems[i].LinkedItems) > 0 {
			for _, linkedItem := range receipt.ReceiptItems[i].LinkedItems {
				linkedItemsData = append(linkedItemsData, LinkedItemData{
					ParentItemIndex: i,
					LinkedItem:      linkedItem,
				})
			}
			// Clear linked items from the receipt item for initial creation
			receipt.ReceiptItems[i].LinkedItems = []models.Item{}
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		repository.SetTransaction(tx)
		notificationRepository.SetTransaction(tx)

		// First nested transaction: Create receipt without linked items
		err := tx.Transaction(func(tx2 *gorm.DB) error {
			err := tx2.Model(models.Receipt{}).Select("*").Create(&receipt).Error
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Second nested transaction: Handle linked items
		if len(linkedItemsData) > 0 {
			err = tx.Transaction(func(tx3 *gorm.DB) error {
				for _, linkedData := range linkedItemsData {
					// Set the receipt ID for the linked item
					linkedData.LinkedItem.ReceiptId = receipt.ID

					// Create the linked item
					err := tx3.Model(models.Item{}).Create(&linkedData.LinkedItem).Error
					if err != nil {
						return err
					}

					// Handle linked item's categories
					if len(linkedData.LinkedItem.Categories) > 0 {
						err = tx3.Model(&linkedData.LinkedItem).Association("Categories").Replace(&linkedData.LinkedItem.Categories)
						if err != nil {
							return err
						}
					}

					// Handle linked item's tags
					if len(linkedData.LinkedItem.Tags) > 0 {
						err = tx3.Model(&linkedData.LinkedItem).Association("Tags").Replace(&linkedData.LinkedItem.Tags)
						if err != nil {
							return err
						}
					}

					// Update the parent item's LinkedItems association
					parentItem := &receipt.ReceiptItems[linkedData.ParentItemIndex]
					err = tx3.Model(parentItem).Association("LinkedItems").Append(&linkedData.LinkedItem)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		var userIdsToOmit []interface{} = make([]interface{}, 1)
		userIdsToOmit = append(userIdsToOmit, *receipt.CreatedBy)

		notificationBody := fmt.Sprintf("The receipt: %s has been uploaded to the group %s. Check it out! %s", receipt.Name, BuildParamaterisedString("groupId", receipt.GroupId, "name", "string"), BuildParamaterisedString("receiptId", receipt.ID, "", "link"))
		err = notificationRepository.SendNotificationToGroup(receipt.GroupId, "Receipt Uploaded", notificationBody, models.NOTIFICATION_TYPE_NORMAL, userIdsToOmit)
		if err != nil {
			return err
		}

		err = repository.AfterReceiptUpdated(&receipt)
		if err != nil {
			return err
		}

		repository.ClearTransaction()
		notificationRepository.ClearTransaction()
		return nil
	})
	if err != nil {
		if !createSystemTask {
			createFailedUpdateSystemTask(systemTask, err)
		}
		return models.Receipt{}, err
	}

	var fullyLoadedReceipt models.Receipt
	err = db.Model(models.Receipt{}).Where("id = ?", receipt.ID).Preload(clause.Associations).Find(&fullyLoadedReceipt).Error
	if err != nil {
		if !createSystemTask {
			createFailedUpdateSystemTask(systemTask, err)
		}
		return models.Receipt{}, err
	}

	if createSystemTask {
		endedAt := time.Now()
		systemTask.EndedAt = &endedAt
		systemTask.AssociatedEntityId = fullyLoadedReceipt.ID
		newReceiptString, err := fullyLoadedReceipt.ToString()
		if err != nil {
			return models.Receipt{}, err
		}

		systemTask.ReceiptId = &fullyLoadedReceipt.ID
		systemTask.GroupId = &fullyLoadedReceipt.GroupId
		systemTask.ResultDescription = newReceiptString

		systemTaskRepository := NewSystemTaskRepository(nil)
		_, err = systemTaskRepository.CreateSystemTask(systemTask)
		if err != nil {
			return models.Receipt{}, err
		}
	}

	return fullyLoadedReceipt, nil
}

func (repository ReceiptRepository) GetReceiptById(receiptId string) (models.Receipt, error) {
	db := GetDB()
	var receipt models.Receipt
	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).First(&receipt).Debug().Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (repository ReceiptRepository) GetPagedReceiptsByGroupId(
	userId uint,
	groupId string,
	pagedRequest commands.ReceiptPagedRequestCommand,
	associations []string,
) ([]models.Receipt, int64, error) {
	var receipts []models.Receipt
	var count int64

	uintGroupId, err := utils.StringToUint(groupId)
	if err != nil {
		return nil, 0, err
	}
	groupRepository := NewGroupRepository(nil)
	isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
	if err != nil {
		return nil, 0, err
	}

	// Apply filter
	query, err := repository.BuildGormFilterQuery(pagedRequest)
	if err != nil {
		return nil, 0, err
	}

	// Filter receipts by group
	if isAllGroup {
		groupMemberRepository := NewGroupMemberRepository(nil)
		groupIds, err := groupMemberRepository.GetGroupIdsByUserId(utils.UintToString(userId))
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("group_id IN ?", groupIds)
	} else {
		query = query.Where("group_id = ?", groupId)
	}

	// Set order by
	if len(pagedRequest.OrderBy) == 0 {
		pagedRequest.OrderBy = "created_at"
	}

	if repository.isTrustedValue(pagedRequest) {
		orderBy := pagedRequest.OrderBy
		query = query.Order(orderBy + " " + string(pagedRequest.SortDirection))
	} else {
		return nil, 0, errors.New("untrusted value " + pagedRequest.OrderBy + " " + string(pagedRequest.SortDirection))
	}

	err = query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	query = query.Preload("Categories").Preload("Tags")

	if pagedRequest.PageSize > 0 && pagedRequest.Page > 0 {
		query = query.Scopes(repository.Paginate(pagedRequest.Page, pagedRequest.PageSize))
	}

	if associations != nil && len(associations) > 0 {
		for _, association := range associations {
			query = query.Preload(association)
		}
	}

	// Run Query
	err = query.Find(&receipts).Error
	if err != nil {
		return nil, 0, err
	}

	return receipts, count, nil
}

func (repository ReceiptRepository) BuildGormFilterQuery(pagedRequest commands.ReceiptPagedRequestCommand) (*gorm.DB, error) {
	query := db.Model(models.Receipt{})
	// Name
	if pagedRequest.Filter.Name.Value != nil {
		name := pagedRequest.Filter.Name.Value.(string)
		if len(name) > 0 {
			query = repository.buildFilterQuery(query, name, pagedRequest.Filter.Name.Operation, "name", false)
		}
	}

	// Date
	if pagedRequest.Filter.Date.Value != nil {
		var date interface{}
		isBetweenOperation := pagedRequest.Filter.Date.Operation == commands.BETWEEN
		if isBetweenOperation {
			date = pagedRequest.Filter.Date.Value.([]interface{})
		} else {
			date = pagedRequest.Filter.Date.Value.(string)
		}

		query = repository.buildFilterQuery(query, date, pagedRequest.Filter.Date.Operation, "date", isBetweenOperation)
	}

	// Paid By
	if pagedRequest.Filter.PaidBy.Value != nil {
		paidBy := pagedRequest.Filter.PaidBy.Value.([]interface{})
		if len(paidBy) > 0 {
			query = repository.buildFilterQuery(query, paidBy, pagedRequest.Filter.PaidBy.Operation, "paid_by_user_id", true)
		}
	}

	// Categories
	if pagedRequest.Filter.Categories.Value != nil {
		categories := pagedRequest.Filter.Categories.Value.([]interface{})
		if len(categories) > 0 {
			if pagedRequest.Filter.Categories.Operation == commands.CONTAINS {
				query = query.Where("id IN (?)", db.Table("receipt_categories").Select("receipt_id").Where("category_id IN ?", categories))
			}
		}

	}

	// Tags
	if pagedRequest.Filter.Tags.Value != nil {
		tags := pagedRequest.Filter.Tags.Value.([]interface{})
		if len(tags) > 0 {
			if pagedRequest.Filter.Tags.Operation == commands.CONTAINS {
				query = query.Where("id IN (?)", db.Table("receipt_tags").Select("receipt_id").Where("tag_id IN ?", tags))
			}
		}
	}

	// Amount
	if pagedRequest.Filter.Amount.Value != nil {
		var amount interface{}
		if pagedRequest.Filter.Amount.Operation == commands.BETWEEN {
			amount = pagedRequest.Filter.Amount.Value.([]interface{})
		} else {
			amount = pagedRequest.Filter.Amount.Value.(float64)
		}
		query = repository.buildFilterQuery(
			query,
			amount,
			pagedRequest.Filter.Amount.Operation,
			"amount", pagedRequest.Filter.Amount.Operation == commands.BETWEEN,
		)
	}

	// Status
	if pagedRequest.Filter.Status.Value != nil {
		status := pagedRequest.Filter.Status.Value.([]interface{})
		if len(status) > 0 {
			query = repository.buildFilterQuery(query, status, pagedRequest.Filter.Status.Operation, "status", true)
		}
	}

	// Resolved Date
	if pagedRequest.Filter.ResolvedDate.Value != nil {
		var resolvedDate interface{}
		isBetweenOperation := pagedRequest.Filter.ResolvedDate.Operation == commands.BETWEEN
		if isBetweenOperation {
			resolvedDate = pagedRequest.Filter.ResolvedDate.Value.(interface{})
		} else {
			resolvedDate = pagedRequest.Filter.ResolvedDate.Value.(string)
		}

		query = repository.buildFilterQuery(
			query,
			resolvedDate,
			pagedRequest.Filter.ResolvedDate.Operation,
			"resolved_date",
			isBetweenOperation,
		)
	}

	// Added At
	if pagedRequest.Filter.CreatedAt.Value != nil {
		var addedAt interface{}
		isBetweenOperation := pagedRequest.Filter.CreatedAt.Operation == commands.BETWEEN
		if isBetweenOperation {
			addedAt = pagedRequest.Filter.CreatedAt.Value.([]interface{})
		} else {
			addedAt = pagedRequest.Filter.CreatedAt.Value.(string)
		}

		query = repository.buildFilterQuery(
			query,
			addedAt,
			pagedRequest.Filter.CreatedAt.Operation,
			"created_at",
			isBetweenOperation,
		)
	}

	return query, nil
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

	if operation == commands.BETWEEN {
		arrayValue := value.([]interface{})
		if len(arrayValue) != 2 {
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

func (repository ReceiptRepository) isTrustedValue(pagedRequest commands.ReceiptPagedRequestCommand) bool {
	orderByTrusted := []interface{}{"date", "name", "paid_by_user_id", "amount", "categories", "tags", "status", "resolved_date", "created_at"}
	directionTrusted := commands.GetValidSortDirections()

	isOrderByTrusted := utils.Contains(orderByTrusted, pagedRequest.OrderBy)
	isDirectionTrusted := utils.Contains(directionTrusted, pagedRequest.SortDirection)

	return isOrderByTrusted && isDirectionTrusted
}

func (repository ReceiptRepository) GetReceiptGroupIdByReceiptId(id string) (uint, error) {
	db := repository.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("group_id").Find(&receipt).Error
	if err != nil {
		return 0, err
	}

	return receipt.GroupId, nil
}

func (repository ReceiptRepository) FilterLinkedItemsFromReceiptItems(receipt *models.Receipt) {
	if len(receipt.ReceiptItems) == 0 {
		return
	}

	// Collect all linked item IDs
	linkedItemIds := make(map[uint]bool)
	for _, item := range receipt.ReceiptItems {
		for _, linkedItem := range item.LinkedItems {
			linkedItemIds[linkedItem.ID] = true
		}
	}

	// Filter out linked items from ReceiptItems
	var filteredItems []models.Item
	for _, item := range receipt.ReceiptItems {
		if !linkedItemIds[item.ID] {
			filteredItems = append(filteredItems, item)
		}
	}

	receipt.ReceiptItems = filteredItems
}

func (repository ReceiptRepository) GetFullyLoadedReceiptById(id string) (models.Receipt, error) {
	db := repository.GetDB()
	var receipt models.Receipt

	query := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations)

	for _, association := range constants.FULL_RECEIPT_ASSOCIATIONS {
		query = query.Preload(association)
	}

	err := query.Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	repository.FilterLinkedItemsFromReceiptItems(&receipt)

	return receipt, nil
}

func (repository ReceiptRepository) GetReceiptsByGroupIds(groupIds []string, querySelect string, queryPreload string) ([]models.Receipt, error) {
	db := repository.GetDB()
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

func (repository ReceiptRepository) GetReceiptsByIds(ids []string, associations []string) ([]models.Receipt, error) {
	query := repository.GetDB().Model(models.Receipt{}).Where("id IN ?", ids)

	hasLinkedItems := false
	if associations != nil {
		for _, association := range associations {
			query = query.Preload(association)
			if association == "ReceiptItems.LinkedItems" {
				hasLinkedItems = true
			}
		}
	}

	var receipts []models.Receipt
	err := query.Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	// Filter linked items if they were loaded
	if hasLinkedItems {
		for i := range receipts {
			repository.FilterLinkedItemsFromReceiptItems(&receipts[i])
		}
	}

	return receipts, nil
}
