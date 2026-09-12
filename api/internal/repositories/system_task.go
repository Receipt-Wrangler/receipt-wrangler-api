package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

// SystemRanByUserId is the sentinel the desktop's "Ran By" picker submits for
// tasks the system ran itself (system_tasks.ran_by_user_id IS NULL). It is
// negative so it can never collide with a real user id — the same convention as
// the role editor's OWN_PAID_RECEIPTS_OPTION_ID and the report builder's
// REPORT_GENERATOR_PAID_BY_ID.
const SystemRanByUserId = -1

type SystemTaskRepository struct {
	BaseRepository
}

func NewSystemTaskRepository(tx *gorm.DB) SystemTaskRepository {
	repository := SystemTaskRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
		TX: tx,
	}}
	return repository
}

func (repository SystemTaskRepository) GetPagedSystemTasks(command commands.GetSystemTaskCommand) ([]models.SystemTask, int64, error) {
	db := repository.GetDB()
	var results []models.SystemTask
	var count int64

	if !isColumnNameValid(command.OrderBy) {
		return nil, 0, errors.New("invalid column name")
	}

	filteredSystemTaskTypes := []models.SystemTaskType{
		models.RECEIPT_UPLOADED,
		models.CHAT_COMPLETION,
		models.OCR_PROCESSING,
	}

	query := db.Model(&models.SystemTask{}).Where("type NOT IN ?", filteredSystemTaskTypes)

	if command.AssociatedEntityId != 0 {
		query = query.Where("associated_entity_id = ?", command.AssociatedEntityId)
	}

	if len(command.AssociatedEntityType) > 0 {
		query = query.Where("associated_entity_type = ?", command.AssociatedEntityType)
	}

	// Applied before Count so totalCount describes the filtered set, mirroring
	// the paid-by/visibility predicates in GetPagedActivities below.
	query = repository.buildSystemTaskFilterQuery(query, command.Filter)

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	err := query.Preload(clause.Associations).Preload("ChildSystemTasks.ChildSystemTasks").Find(&results).Error
	if query.Error != nil {
		return nil, 0, err
	}

	return results, count, nil
}

// buildSystemTaskFilterQuery narrows the paged system task query by the
// caller's filter. Every value arrives as an interface{} off the request body,
// so each unwrap is a comma-ok assertion: a wrong-typed value is treated as
// "field not set" rather than panicking the handler.
func (repository SystemTaskRepository) buildSystemTaskFilterQuery(query *gorm.DB, filter commands.SystemTaskPagedRequestFilter) *gorm.DB {
	if types, ok := filter.Type.Value.([]interface{}); ok && len(types) > 0 {
		query = repository.BuildFilterQuery(query, types, filter.Type.Operation, "type", true)
	}

	query = repository.applyRanByFilter(query, filter.RanBy)
	query = repository.applyTimestampDayFilter(query, filter.StartedAt, "started_at")
	query = repository.applyTimestampDayFilter(query, filter.EndedAt, "ended_at")

	return query
}

// applyRanByFilter cannot go through BuildFilterQuery because ran_by_user_id is
// nullable and the rows with no user are exactly the ones the table labels
// "System". The desktop submits SystemRanByUserId for those, so the sentinel is
// split out of the id list and becomes an IS NULL disjunct.
func (repository SystemTaskRepository) applyRanByFilter(query *gorm.DB, field commands.PagedRequestField) *gorm.DB {
	rawIds, ok := field.Value.([]interface{})
	if !ok || len(rawIds) == 0 || field.Operation != commands.CONTAINS {
		return query
	}

	includeSystem := false
	userIds := make([]int64, 0, len(rawIds))

	for _, rawId := range rawIds {
		id, ok := toInt64(rawId)
		if !ok {
			continue
		}

		if id == SystemRanByUserId {
			includeSystem = true
			continue
		}

		userIds = append(userIds, id)
	}

	switch {
	case includeSystem && len(userIds) > 0:
		// Parenthesized explicitly, like applyActivityVisibilityDisjunction below:
		// GORM does wrap an OR condition when AND-ing it onto the query today, but
		// an unparenthesized disjunction would bind as
		// `(other AND a IS NULL) OR a IN (...)` and silently drop the other filters
		// from the second branch.
		return query.Where("(ran_by_user_id IS NULL OR ran_by_user_id IN ?)", userIds)
	case includeSystem:
		return query.Where("ran_by_user_id IS NULL")
	case len(userIds) > 0:
		return query.Where("ran_by_user_id IN ?", userIds)
	default:
		return query
	}
}

// applyTimestampDayFilter compares a timestamp column against whole calendar
// days. started_at / ended_at carry a time of day, unlike the date-only column
// the receipt filter compares against, so a raw comparison to the datepicker's
// midnight would make EQUALS never match and BETWEEN drop everything after
// midnight on the end day.
//
// "Day" resolves in the server's location, the same zone WITHIN_CURRENT_MONTH
// already uses — a client in a different zone can therefore shift the boundary
// by a day, exactly as it can for receipts.
//
// ended_at is nullable, so any filter on it excludes tasks that are still
// running. That is the correct reading of "ended before X".
func (repository SystemTaskRepository) applyTimestampDayFilter(query *gorm.DB, field commands.PagedRequestField, column string) *gorm.DB {
	if field.Value == nil {
		return query
	}

	if field.Operation == commands.WITHIN_CURRENT_MONTH {
		return repository.BuildFilterQuery(query, field.Value, field.Operation, column, false)
	}

	if field.Operation == commands.BETWEEN {
		bounds, ok := field.Value.([]interface{})
		if !ok || len(bounds) != 2 {
			return query
		}

		start, startOk := startOfDayValue(bounds[0])
		end, endOk := startOfDayValue(bounds[1])
		if !startOk || !endOk {
			return query
		}

		return query.Where(column+" >= ? AND "+column+" < ?", start, end.AddDate(0, 0, 1))
	}

	day, ok := startOfDayValue(field.Value)
	if !ok {
		return query
	}

	switch field.Operation {
	case commands.EQUALS:
		return query.Where(column+" >= ? AND "+column+" < ?", day, day.AddDate(0, 0, 1))
	case commands.GREATER_THAN:
		return query.Where(column+" >= ?", day.AddDate(0, 0, 1))
	case commands.LESS_THAN:
		return query.Where(column+" < ?", day)
	default:
		return query
	}
}

// startOfDayValue parses a filter value into the midnight that begins its
// calendar day in the server's location. Values ride the wire as the ISO
// strings Date.toJSON() produces; a time.Time is accepted so Go callers and
// tests can pass one directly.
func startOfDayValue(value interface{}) (time.Time, bool) {
	var parsed time.Time

	switch typed := value.(type) {
	case time.Time:
		parsed = typed
	case string:
		if len(typed) == 0 {
			return time.Time{}, false
		}

		var err error
		parsed, err = time.Parse(time.RFC3339, typed)
		if err != nil {
			// A bare calendar day (yyyy-MM-dd) is already midnight-local.
			parsed, err = time.ParseInLocation(time.DateOnly, typed, time.Local)
			if err != nil {
				return time.Time{}, false
			}
		}
	default:
		return time.Time{}, false
	}

	local := parsed.In(time.Local)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local), true
}

// toInt64 normalizes a JSON-decoded number. Ids arrive as float64 through
// encoding/json, but a Go caller may pass a native integer type.
func toInt64(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case float32:
		return int64(typed), true
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	default:
		return 0, false
	}
}

// ActivityVisibilityResolver reports, for a group, the ran-by user ids the caller may
// see under member isolation, or unrestricted == true (see every actor). It is the
// activity analogue of PaidByAllowedResolver: it lets the handler inject the service-layer
// per-group visible set without the repository importing services. A nil resolver skips
// isolation filtering (backward compatible).
type ActivityVisibilityResolver func(groupId uint) (visibleUserIds []uint, unrestricted bool, err error)

func (repository SystemTaskRepository) GetPagedActivities(
	command commands.PagedActivityRequestCommand,
	resolver ActivityVisibilityResolver,
) (
	[]structs.Activity,
	int64,
	error,
) {
	db := repository.GetDB()
	var results []structs.Activity
	var count int64

	if !isColumnNameValid(command.OrderBy) {
		return nil, 0, errors.New("invalid column name")
	}

	systemTaskTypesToGet := []models.SystemTaskType{
		models.QUICK_SCAN,
		models.RECEIPT_UPLOADED,
		models.RECEIPT_UPDATED,
		models.EMAIL_UPLOAD,
	}

	query := db.Model(&models.SystemTask{}).
		Omit("can_be_restarted").
		Where("type IN ?", systemTaskTypesToGet).
		Where("group_id IN ?", command.GroupIds).
		Not(db.Where("type = ? AND ran_by_user_id IS NULL", models.RECEIPT_UPLOADED))

	// Member isolation: drop activities run by a user the caller may not see in that
	// activity's group IN THE QUERY (before Count + pagination), so TotalCount and the
	// returned page both reflect only visible rows and DB-side LIMIT/OFFSET is preserved.
	query, err := repository.applyActivityVisibilityDisjunction(query, command.GroupIds, resolver)
	if err != nil {
		return nil, 0, err
	}

	query.Count(&count)

	query = repository.Sort(query, command.OrderBy, command.SortDirection)
	query = query.Scopes(repository.Paginate(command.Page, command.PageSize))

	query.Find(&results)

	return results, count, nil
}

// applyActivityVisibilityDisjunction AND-s a per-group actor-visibility disjunction onto
// the query, mirroring ReceiptRepository.ApplyPaidByDisjunction. A nil resolver adds no
// predicate. For each group the caller sees every actor (unrestricted) the clause is just
// group_id = G; otherwise it is group_id = G AND (ran_by_user_id IS NULL OR ran_by_user_id
// IN <visible ids>) — so system actions (nil ran-by) stay visible and only visible
// members' activities survive. Applied before Count so pagination stays consistent.
func (repository SystemTaskRepository) applyActivityVisibilityDisjunction(
	query *gorm.DB,
	groupIds []uint,
	resolver ActivityVisibilityResolver,
) (*gorm.DB, error) {
	if resolver == nil {
		return query, nil
	}
	if len(groupIds) == 0 {
		return query.Where("1 = 0"), nil
	}

	disjunction := repository.GetDB().Session(&gorm.Session{NewDB: true})
	for _, groupId := range groupIds {
		visibleIds, unrestricted, err := resolver(groupId)
		if err != nil {
			return nil, err
		}
		if unrestricted {
			disjunction = disjunction.Or("group_id = ?", groupId)
		} else {
			groupCondition := repository.GetDB().Session(&gorm.Session{NewDB: true}).
				Where("group_id = ?", groupId).
				Where("(ran_by_user_id IS NULL OR ran_by_user_id IN ?)", activityInValues(visibleIds))
			disjunction = disjunction.Or(groupCondition)
		}
	}

	return query.Where(disjunction), nil
}

// activityInValues guards the IN clause against an empty visible set (a restricted set
// always contains at least the caller's own id, but guard defensively): ran-by user ids
// start at 1, so 0 matches no row, yielding "see nothing" rather than a malformed IN ().
func activityInValues(visibleUserIds []uint) []uint {
	if len(visibleUserIds) == 0 {
		return []uint{0}
	}
	return visibleUserIds
}

func isColumnNameValid(columnName string) bool {
	return columnName == "type" || columnName == "status" || columnName == "associated_entity_type" || columnName == "associated_entity_id" || columnName == "started_at" || columnName == "ended_at" || columnName == "result_description" || columnName == "ran_by_user_id"
}

func (repository SystemTaskRepository) CreateSystemTask(command commands.UpsertSystemTaskCommand) (models.SystemTask, error) {
	db := repository.GetDB()

	systemTask := models.SystemTask{
		Type:                   command.Type,
		Status:                 command.Status,
		AssociatedEntityType:   command.AssociatedEntityType,
		AssociatedEntityId:     command.AssociatedEntityId,
		StartedAt:              command.StartedAt,
		EndedAt:                command.EndedAt,
		ResultDescription:      command.ResultDescription,
		RanByUserId:            command.RanByUserId,
		ReceiptId:              command.ReceiptId,
		GroupId:                command.GroupId,
		AssociatedSystemTaskId: command.AssociatedSystemTaskId,
		AsynqTaskId:            command.AsynqTaskId,
	}

	err := db.Create(&systemTask).Error
	if err != nil {
		return models.SystemTask{}, err
	}

	if command.AssociatedSystemTaskId != nil && systemTask.Status == models.SYSTEM_TASK_FAILED {
		var parentSystemTask models.SystemTask
		db.Model(&models.SystemTask{}).Where("id = ?", command.AssociatedSystemTaskId).Find(&parentSystemTask)

		if parentSystemTask.Status == models.SYSTEM_TASK_SUCCEEDED {
			db.Model(&parentSystemTask).Update("status", models.SYSTEM_TASK_FAILED)
		}

	}

	return systemTask, nil
}

func (repository SystemTaskRepository) DeleteSystemTaskByAssociatedEntityId(
	associatedEntityId string,
	emailType models.AssociatedEntityType,
) error {
	db := repository.GetDB()
	err := db.Where("associated_entity_id = ? and associated_entity_type = ?", associatedEntityId, emailType).Delete(&models.SystemTask{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository SystemTaskRepository) GetSystemTaskById(id uint) (models.SystemTask, error) {
	db := repository.GetDB()
	var systemTask models.SystemTask

	err := db.Model(&models.SystemTask{}).Where("id = ?", id).First(&systemTask).Error
	if err != nil {
		return models.SystemTask{}, err
	}

	return systemTask, nil
}

func (repository SystemTaskRepository) AssociateSystemTaskToReceipt(receiptId uint, systemTaskId uint) error {
	db := repository.GetDB()
	return db.Model(&models.SystemTask{}).Where("id = ?", systemTaskId).Update("receipt_id", receiptId).Error
}
