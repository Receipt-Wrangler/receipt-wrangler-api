package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"testing"
	"time"

	"gorm.io/gorm"
)

// GetPagedActivities applies the actor-visibility disjunction in SQL BEFORE COUNT/LIMIT:
// a hidden actor's activities affect neither the returned rows nor the total, even when
// they are newer and would otherwise fill the first page.
func TestGetPagedActivitiesFiltersByVisibilityBeforeCountAndLimit(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	group := models.Group{Name: "act-vis-grp"}
	db.Create(&group)
	visible := models.User{Username: "act-visible", Password: "x"}
	hidden := models.User{Username: "act-hidden", Password: "x"}
	db.Create(&visible)
	db.Create(&hidden)

	gid := group.ID
	// Hidden activities are NEWER (sort first by started_at desc); the visible one is the
	// oldest, so a limit applied before filtering would drop it.
	for i := 0; i < 5; i++ {
		hid := hidden.ID
		db.Create(&models.SystemTask{
			Type:                 models.QUICK_SCAN,
			Status:               models.SYSTEM_TASK_SUCCEEDED,
			AssociatedEntityType: models.NOOP_ENTITY_TYPE,
			GroupId:              &gid,
			RanByUserId:          &hid,
			StartedAt:            time.Now().Add(time.Duration(i+1) * time.Hour),
		})
	}
	vid := visible.ID
	db.Create(&models.SystemTask{
		Type:                 models.QUICK_SCAN,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.NOOP_ENTITY_TYPE,
		GroupId:              &gid,
		RanByUserId:          &vid,
		StartedAt:            time.Now(),
	})

	// The resolver restricts the group to the visible actor only.
	resolver := func(groupId uint) ([]uint, bool, error) {
		return []uint{visible.ID}, false, nil
	}

	command := commands.PagedActivityRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      3,
			OrderBy:       "started_at",
			SortDirection: commands.DESCENDING,
		},
		GroupIds: []uint{group.ID},
	}

	activities, count, err := NewSystemTaskRepository(nil).GetPagedActivities(command, resolver)
	if err != nil {
		t.Fatalf("GetPagedActivities: %v", err)
	}

	// Only the visible actor's one activity is counted and returned, even though the
	// hidden ones are newer and a pre-filter limit of 3 would have returned them.
	if count != 1 {
		t.Errorf("TotalCount = %d, want 1 (filtered before count)", count)
	}
	if len(activities) != 1 {
		t.Fatalf("returned %d rows, want 1", len(activities))
	}
	if activities[0].RanByUserId == nil || *activities[0].RanByUserId != visible.ID {
		t.Errorf("returned activity should be the visible actor's, got %+v", activities[0])
	}
}

// A multi-group request combining an unrestricted (open) group and a restricted (isolated)
// group exercises the `unrestricted`-per-group OR branch: the open group's hidden-actor
// activity is kept while the isolated group's hidden-actor activity is dropped.
func TestGetPagedActivitiesMixedVisibilityAcrossGroups(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	openGroup := models.Group{Name: "act-open-grp"}
	isolatedGroup := models.Group{Name: "act-isolated-grp"}
	db.Create(&openGroup)
	db.Create(&isolatedGroup)
	visible := models.User{Username: "act-mix-visible", Password: "x"}
	hidden := models.User{Username: "act-mix-hidden", Password: "x"}
	db.Create(&visible)
	db.Create(&hidden)

	oid, iid, hid, vid := openGroup.ID, isolatedGroup.ID, hidden.ID, visible.ID

	newTask := func(groupId, ranBy uint) *models.SystemTask {
		gid, uid := groupId, ranBy
		return &models.SystemTask{
			Type:                 models.QUICK_SCAN,
			Status:               models.SYSTEM_TASK_SUCCEEDED,
			AssociatedEntityType: models.NOOP_ENTITY_TYPE,
			GroupId:              &gid,
			RanByUserId:          &uid,
			StartedAt:            time.Now(),
		}
	}
	// Open group: the hidden user's activity should still show (unrestricted).
	db.Create(newTask(oid, hid))
	// Isolated group: the hidden user's activity is dropped; the visible user's is kept.
	db.Create(newTask(iid, hid))
	db.Create(newTask(iid, vid))

	resolver := func(groupId uint) ([]uint, bool, error) {
		if groupId == oid {
			return nil, true, nil // unrestricted for the open group
		}
		return []uint{vid}, false, nil // restricted to the visible user in the isolated group
	}

	command := commands.PagedActivityRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "started_at",
			SortDirection: commands.DESCENDING,
		},
		GroupIds: []uint{oid, iid},
	}

	activities, count, err := NewSystemTaskRepository(nil).GetPagedActivities(command, resolver)
	if err != nil {
		t.Fatalf("GetPagedActivities: %v", err)
	}

	// Open-group hidden actor (kept) + isolated-group visible actor (kept) = 2; the
	// isolated-group hidden actor is dropped.
	if count != 2 {
		t.Errorf("TotalCount = %d, want 2", count)
	}
	if len(activities) != 2 {
		t.Fatalf("returned %d rows, want 2", len(activities))
	}
	for _, a := range activities {
		if a.GroupId != nil && *a.GroupId == iid && a.RanByUserId != nil && *a.RanByUserId == hid {
			t.Errorf("hidden actor's isolated-group activity should be dropped")
		}
	}
}

// A nil resolver adds no predicate (backward compatible): every matching activity is
// returned and counted.
func TestGetPagedActivitiesNilResolverUnfiltered(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	group := models.Group{Name: "act-nil-grp"}
	db.Create(&group)
	user := models.User{Username: "act-nil-user", Password: "x"}
	db.Create(&user)

	gid := group.ID
	uid := user.ID
	for i := 0; i < 2; i++ {
		db.Create(&models.SystemTask{
			Type:                 models.QUICK_SCAN,
			Status:               models.SYSTEM_TASK_SUCCEEDED,
			AssociatedEntityType: models.NOOP_ENTITY_TYPE,
			GroupId:              &gid,
			RanByUserId:          &uid,
			StartedAt:            time.Now(),
		})
	}

	command := commands.PagedActivityRequestCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      10,
			OrderBy:       "started_at",
			SortDirection: commands.DESCENDING,
		},
		GroupIds: []uint{group.ID},
	}

	activities, count, err := NewSystemTaskRepository(nil).GetPagedActivities(command, nil)
	if err != nil {
		t.Fatalf("GetPagedActivities: %v", err)
	}
	if count != 2 || len(activities) != 2 {
		t.Errorf("nil resolver should return all rows unfiltered, got count=%d rows=%d", count, len(activities))
	}
}

// --- System task filter -------------------------------------------------
//
// Every case below drives GetPagedSystemTasks end to end so it also proves the
// predicates land BEFORE Count: totalCount is asserted alongside the rows.

func seedSystemTask(db *gorm.DB, taskType models.SystemTaskType, ranBy *uint, startedAt time.Time, endedAt *time.Time) models.SystemTask {
	task := models.SystemTask{
		Type:                 taskType,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.NOOP_ENTITY_TYPE,
		RanByUserId:          ranBy,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
	}
	db.Create(&task)

	return task
}

func systemTaskFilterCommand(filter commands.SystemTaskPagedRequestFilter) commands.GetSystemTaskCommand {
	return commands.GetSystemTaskCommand{
		PagedRequestCommand: commands.PagedRequestCommand{
			Page:          1,
			PageSize:      50,
			OrderBy:       "started_at",
			SortDirection: commands.DESCENDING,
		},
		Filter: filter,
	}
}

// The three child-only types are never returned as top-level rows, which is why
// the desktop's Type picker omits them (CHILD_ONLY_SYSTEM_TASK_TYPES in
// desktop/src/constants/system-task-type-options.ts). Keep the two in sync.
func TestGetPagedSystemTasksExcludesChildTaskTypes(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	now := time.Now()

	for _, childType := range []models.SystemTaskType{
		models.RECEIPT_UPLOADED,
		models.CHAT_COMPLETION,
		models.OCR_PROCESSING,
	} {
		seedSystemTask(db, childType, nil, now, nil)
	}
	seedSystemTask(db, models.QUICK_SCAN, nil, now, nil)

	repository := NewSystemTaskRepository(nil)
	results, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 1 || len(results) != 1 || results[0].Type != models.QUICK_SCAN {
		t.Errorf("expected only the QUICK_SCAN task, got count=%d rows=%d", count, len(results))
	}
}

func TestGetPagedSystemTasksFiltersByType(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	now := time.Now()

	seedSystemTask(db, models.QUICK_SCAN, nil, now, nil)
	seedSystemTask(db, models.MAGIC_FILL, nil, now, nil)
	seedSystemTask(db, models.EMAIL_READ, nil, now, nil)

	repository := NewSystemTaskRepository(nil)
	results, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		Type: commands.PagedRequestField{
			Operation: commands.CONTAINS,
			Value:     []interface{}{string(models.QUICK_SCAN), string(models.EMAIL_READ)},
		},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 || len(results) != 2 {
		t.Fatalf("expected 2 rows and a total of 2, got count=%d rows=%d", count, len(results))
	}

	for _, result := range results {
		if result.Type != models.QUICK_SCAN && result.Type != models.EMAIL_READ {
			t.Errorf("unexpected type %v in results", result.Type)
		}
	}
}

func TestGetPagedSystemTasksFiltersByRanBy(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	now := time.Now()

	alice := models.User{Username: "task-filter-alice", Password: "x"}
	bob := models.User{Username: "task-filter-bob", Password: "x"}
	db.Create(&alice)
	db.Create(&bob)

	aliceId := alice.ID
	bobId := bob.ID
	seedSystemTask(db, models.QUICK_SCAN, &aliceId, now, nil)
	seedSystemTask(db, models.QUICK_SCAN, &bobId, now, nil)
	// Two system-run tasks: the rows the table labels "System".
	seedSystemTask(db, models.QUICK_SCAN, nil, now, nil)
	seedSystemTask(db, models.EMAIL_READ, nil, now, nil)

	tests := map[string]struct {
		value    []interface{}
		expected int64
	}{
		// Ids arrive as float64 through encoding/json, matching production.
		"ids only":             {value: []interface{}{float64(aliceId)}, expected: 1},
		"system sentinel only": {value: []interface{}{float64(SystemRanByUserId)}, expected: 2},
		"sentinel and ids":     {value: []interface{}{float64(SystemRanByUserId), float64(bobId)}, expected: 3},
	}

	repository := NewSystemTaskRepository(nil)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			results, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
				RanBy: commands.PagedRequestField{
					Operation: commands.CONTAINS,
					Value:     test.value,
				},
			}))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if count != test.expected || int64(len(results)) != test.expected {
				t.Errorf("expected %d rows and a total of %d, got count=%d rows=%d",
					test.expected, test.expected, count, len(results))
			}
		})
	}
}

// The headline whole-day case, asserted on the bound values rather than on
// rows: started_at carries a time of day, so a raw comparison against the
// datepicker's midnight would drop a task that ran late on the last day of the
// range. It has to be asserted this way because the test DB is SQLite, which
// stores timestamps as text and compares them lexically — there, a raw bound
// happens to sort *after* every row on the same day (" " < "T"), so the row
// count alone cannot tell the widened path from the broken one. A real
// Postgres/MySQL deployment casts to a timestamp and drops those rows.
func TestApplyTimestampDayFilterWidensBoundsToWholeDays(t *testing.T) {
	defer TruncateTestDb()

	repository := NewSystemTaskRepository(nil)
	day := time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local)
	nextDay := day.AddDate(0, 0, 1)

	tests := map[string]struct {
		field    commands.PagedRequestField
		expected []time.Time
	}{
		"equals covers the whole day": {
			field:    commands.PagedRequestField{Operation: commands.EQUALS, Value: day.Format(time.RFC3339)},
			expected: []time.Time{day, nextDay},
		},
		"greater than starts at the next day": {
			field:    commands.PagedRequestField{Operation: commands.GREATER_THAN, Value: day.Format(time.RFC3339)},
			expected: []time.Time{nextDay},
		},
		"less than stops at the start of the day": {
			field:    commands.PagedRequestField{Operation: commands.LESS_THAN, Value: day.Format(time.RFC3339)},
			expected: []time.Time{day},
		},
		"between runs to the end of the last day": {
			field: commands.PagedRequestField{
				Operation: commands.BETWEEN,
				Value: []interface{}{
					time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local).Format(time.RFC3339),
					day.Format(time.RFC3339),
				},
			},
			expected: []time.Time{time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local), nextDay},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			query := GetDB().Session(&gorm.Session{DryRun: true}).Model(&models.SystemTask{})
			query = repository.applyTimestampDayFilter(query, test.field, "started_at")

			var results []models.SystemTask
			statement := query.Find(&results).Statement

			bounds := []time.Time{}
			for _, variable := range statement.Vars {
				if bound, ok := variable.(time.Time); ok {
					bounds = append(bounds, bound)
				}
			}

			if len(bounds) != len(test.expected) {
				t.Fatalf("expected %d bounds, got %d (%v) -- SQL: %s",
					len(test.expected), len(bounds), bounds, statement.SQL.String())
			}

			for i, expected := range test.expected {
				if !bounds[i].Equal(expected) {
					t.Errorf("bound %d: expected %v, got %v", i, expected, bounds[i])
				}
			}
		})
	}
}

func TestGetPagedSystemTasksBetweenIncludesTasksLateOnTheEndDay(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	rangeStart := time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local)
	rangeEnd := time.Date(2026, 3, 12, 0, 0, 0, 0, time.Local)

	lateOnEndDay := seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 12, 23, 30, 0, 0, time.Local), nil)
	insideRange := seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local), nil)
	seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 13, 0, 30, 0, 0, time.Local), nil)
	seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 9, 23, 30, 0, 0, time.Local), nil)

	repository := NewSystemTaskRepository(nil)
	results, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		StartedAt: commands.PagedRequestField{
			Operation: commands.BETWEEN,
			Value: []interface{}{
				rangeStart.Format(time.RFC3339),
				rangeEnd.Format(time.RFC3339),
			},
		},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 || len(results) != 2 {
		t.Fatalf("expected 2 rows and a total of 2, got count=%d rows=%d", count, len(results))
	}

	returned := map[uint]bool{}
	for _, result := range results {
		returned[result.ID] = true
	}

	if !returned[lateOnEndDay.ID] {
		t.Error("expected the task that ran late on the end day to be included")
	}
	if !returned[insideRange.ID] {
		t.Error("expected the task inside the range to be included")
	}
}

func TestGetPagedSystemTasksEqualsMatchesTheWholeDay(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	day := time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local)

	seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 11, 0, 0, 1, 0, time.Local), nil)
	seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 11, 23, 59, 59, 0, time.Local), nil)
	seedSystemTask(db, models.QUICK_SCAN, nil, time.Date(2026, 3, 12, 0, 0, 1, 0, time.Local), nil)

	repository := NewSystemTaskRepository(nil)
	_, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		StartedAt: commands.PagedRequestField{
			Operation: commands.EQUALS,
			Value:     day.Format(time.RFC3339),
		},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 {
		t.Errorf("expected the 2 tasks that started on that calendar day, got %d", count)
	}
}

// ended_at is nullable, so a filter on it also excludes tasks that never ended.
func TestGetPagedSystemTasksLessThanOnEndedAtExcludesRunningTasks(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()

	started := time.Date(2026, 3, 10, 8, 0, 0, 0, time.Local)
	endedEarlier := time.Date(2026, 3, 10, 9, 0, 0, 0, time.Local)
	endedOnCutoffDay := time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local)

	seedSystemTask(db, models.QUICK_SCAN, nil, started, &endedEarlier)
	seedSystemTask(db, models.QUICK_SCAN, nil, started, &endedOnCutoffDay)
	seedSystemTask(db, models.QUICK_SCAN, nil, started, nil)

	repository := NewSystemTaskRepository(nil)
	_, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		EndedAt: commands.PagedRequestField{
			Operation: commands.LESS_THAN,
			Value:     time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local).Format(time.RFC3339),
		},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 1 {
		t.Errorf("expected only the task that ended before the cutoff day, got %d", count)
	}
}

// A malformed body must be a query no-op, not a panic: every unwrap in
// buildSystemTaskFilterQuery is a comma-ok assertion.
func TestGetPagedSystemTasksIgnoresWrongTypedFilterValues(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	now := time.Now()

	seedSystemTask(db, models.QUICK_SCAN, nil, now, nil)
	seedSystemTask(db, models.MAGIC_FILL, nil, now, nil)

	repository := NewSystemTaskRepository(nil)
	_, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		Type:      commands.PagedRequestField{Operation: commands.CONTAINS, Value: "QUICK_SCAN"},
		RanBy:     commands.PagedRequestField{Operation: commands.CONTAINS, Value: float64(1)},
		StartedAt: commands.PagedRequestField{Operation: commands.EQUALS, Value: "not-a-date"},
		EndedAt:   commands.PagedRequestField{Operation: commands.BETWEEN, Value: []interface{}{"nope"}},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 {
		t.Errorf("expected an unfiltered result of 2, got %d", count)
	}
}

// A ranBy filter that includes the System sentinel is a disjunction, so it has
// to be parenthesized: unparenthesized, `type IN (..) AND a IS NULL OR a IN (..)`
// binds as `(type AND a IS NULL) OR a IN (..)` and the type filter silently
// stops applying to the second branch.
func TestGetPagedSystemTasksCombinesTypeWithASystemRanByDisjunction(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	now := time.Now()

	alice := models.User{Username: "task-filter-combo", Password: "x"}
	db.Create(&alice)
	aliceId := alice.ID

	seedSystemTask(db, models.QUICK_SCAN, &aliceId, now, nil) // matches both
	seedSystemTask(db, models.QUICK_SCAN, nil, now, nil)      // matches both
	seedSystemTask(db, models.MAGIC_FILL, &aliceId, now, nil) // wrong type
	seedSystemTask(db, models.MAGIC_FILL, nil, now, nil)      // wrong type

	repository := NewSystemTaskRepository(nil)
	_, count, err := repository.GetPagedSystemTasks(systemTaskFilterCommand(commands.SystemTaskPagedRequestFilter{
		Type: commands.PagedRequestField{
			Operation: commands.CONTAINS,
			Value:     []interface{}{string(models.QUICK_SCAN)},
		},
		RanBy: commands.PagedRequestField{
			Operation: commands.CONTAINS,
			Value:     []interface{}{float64(SystemRanByUserId), float64(aliceId)},
		},
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 {
		t.Errorf("expected only the 2 QUICK_SCAN tasks, got %d", count)
	}
}
