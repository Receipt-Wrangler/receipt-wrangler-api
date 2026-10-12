package repositories

import (
	"strings"
	"sync"
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"

	"gorm.io/gorm"
)

func intPtr(value int) *int {
	return &value
}

// buildSettingsCommand returns a command that passes validation, with both
// lifetimes omitted by default so callers can set only the one under test.
func buildSettingsCommand() commands.UpsertSystemSettingsCommand {
	queueConfigs := make([]commands.UpsertTaskQueueConfigurationCommand, 0)
	for _, config := range models.GetAllDefaultQueueConfigurations() {
		queueConfigs = append(queueConfigs, commands.UpsertTaskQueueConfigurationCommand{
			Name:     config.Name,
			Priority: 1,
		})
	}

	return commands.UpsertSystemSettingsCommand{
		CurrencyDisplay:              "$",
		CurrencySymbolPosition:       models.START,
		CurrencyThousandthsSeparator: models.COMMA,
		CurrencyDecimalSeparator:     models.DOT,
		TaskConcurrency:              1,
		EmailPollingInterval:         60,
		TaskQueueConfigurations:      queueConfigs,
	}
}

func seedLifetimes(t *testing.T, appHours int, mcpHours int) {
	t.Helper()

	if err := GetDB().Create(&models.SystemSettings{}).Error; err != nil {
		t.Fatalf("failed to create system settings: %v", err)
	}

	err := GetDB().Model(&models.SystemSettings{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"refresh_token_valid_for_hours":     appHours,
			"mcp_refresh_token_valid_for_hours": mcpHours,
		}).Error
	if err != nil {
		t.Fatalf("failed to seed lifetimes: %v", err)
	}
}

// A lifetime the request omitted must be left out of the UPDATE statement
// altogether. Copying the stored value onto the row instead would still write
// the column, which is what lets a concurrent update clobber it. Asserted on the
// generated SQL because that is the precise property the concurrency safety
// rests on.
func TestUpdateSystemSettingsOmitsUnsentLifetimeColumns(t *testing.T) {
	defer TruncateTestDb()
	seedLifetimes(t, 24, 24)

	settings := models.SystemSettings{}
	settings.ID = 1

	tests := map[string]struct {
		command       commands.UpsertSystemSettingsCommand
		expectWritten []string
		expectSkipped []string
	}{
		"both omitted": {
			command: buildSettingsCommand(),
			expectSkipped: []string{
				"`refresh_token_valid_for_hours`",
				"`mcp_refresh_token_valid_for_hours`",
				"`temp_file_retention_hours`",
			},
		},
		"retention sent, lifetimes omitted": {
			command: func() commands.UpsertSystemSettingsCommand {
				cmd := buildSettingsCommand()
				cmd.TempFileRetentionHours = intPtr(1080)
				return cmd
			}(),
			expectWritten: []string{"`temp_file_retention_hours`"},
			expectSkipped: []string{"`refresh_token_valid_for_hours`", "`mcp_refresh_token_valid_for_hours`"},
		},
		"lifetimes sent, retention omitted": {
			command: func() commands.UpsertSystemSettingsCommand {
				cmd := buildSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(720)
				cmd.McpRefreshTokenValidForHours = intPtr(6)
				return cmd
			}(),
			expectWritten: []string{"`refresh_token_valid_for_hours`", "`mcp_refresh_token_valid_for_hours`"},
			expectSkipped: []string{"`temp_file_retention_hours`"},
		},
		"app sent, mcp omitted": {
			command: func() commands.UpsertSystemSettingsCommand {
				cmd := buildSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(720)
				return cmd
			}(),
			expectWritten: []string{"`refresh_token_valid_for_hours`"},
			expectSkipped: []string{"`mcp_refresh_token_valid_for_hours`"},
		},
		"mcp sent, app omitted": {
			command: func() commands.UpsertSystemSettingsCommand {
				cmd := buildSettingsCommand()
				cmd.McpRefreshTokenValidForHours = intPtr(6)
				return cmd
			}(),
			expectWritten: []string{"`mcp_refresh_token_valid_for_hours`"},
			expectSkipped: []string{"`refresh_token_valid_for_hours`"},
		},
		"both sent": {
			command: func() commands.UpsertSystemSettingsCommand {
				cmd := buildSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(720)
				cmd.McpRefreshTokenValidForHours = intPtr(6)
				return cmd
			}(),
			expectWritten: []string{"`refresh_token_valid_for_hours`", "`mcp_refresh_token_valid_for_hours`"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			omitted := append([]string{"TaskQueueConfigurations"}, test.command.OmittedColumns()...)
			sql := GetDB().Session(&gorm.Session{DryRun: true}).
				Model(&settings).Select("*").Omit(omitted...).
				Where("id = ?", 1).Updates(&settings).Statement.SQL.String()

			for _, column := range test.expectWritten {
				if !strings.Contains(sql, column+"=") {
					t.Errorf("expected %s to be written, SQL: %s", column, sql)
				}
			}

			for _, column := range test.expectSkipped {
				if strings.Contains(sql, column+"=") {
					t.Errorf("expected %s to be skipped, SQL: %s", column, sql)
				}
			}
		})
	}
}

// Two admins saving at once, each setting one lifetime and omitting the other.
// Both explicit values must survive: with the columns skipped there is nothing
// for either write to clobber. Repeated because the losing interleaving is
// timing dependent -- a single pass could pass by luck against a broken
// implementation.
func TestUpdateSystemSettingsConcurrentLifetimeUpdatesBothSurvive(t *testing.T) {
	defer TruncateTestDb()

	const attempts = 15

	for attempt := 0; attempt < attempts; attempt++ {
		TruncateTestDb()
		seedLifetimes(t, 24, 24)

		appCommand := buildSettingsCommand()
		appCommand.RefreshTokenValidForHours = intPtr(720)

		mcpCommand := buildSettingsCommand()
		mcpCommand.McpRefreshTokenValidForHours = intPtr(6)

		var start sync.WaitGroup
		var done sync.WaitGroup
		start.Add(1)

		for _, command := range []commands.UpsertSystemSettingsCommand{appCommand, mcpCommand} {
			done.Add(1)
			go func(c commands.UpsertSystemSettingsCommand) {
				defer done.Done()
				start.Wait()
				NewSystemSettingsRepository(nil).UpdateSystemSettings(c)
			}(command)
		}

		start.Done()
		done.Wait()

		settings, err := NewSystemSettingsRepository(nil).GetSystemSettings()
		if err != nil {
			t.Fatalf("failed to read system settings: %v", err)
		}

		if settings.RefreshTokenValidForHours != 720 {
			t.Fatalf("attempt %d: app lifetime = %d, expected 720 (clobbered by the concurrent MCP update)", attempt, settings.RefreshTokenValidForHours)
		}

		if settings.McpRefreshTokenValidForHours != 6 {
			t.Fatalf("attempt %d: mcp lifetime = %d, expected 6 (clobbered by the concurrent app update)", attempt, settings.McpRefreshTokenValidForHours)
		}
	}
}

// seedFourQueueConfigurations reproduces an install that last saved its settings
// before models.SystemCleanUpQueue existed: it has a row for every other queue and
// none for that one. Priorities are distinctive so a default silently overwriting a
// persisted value is visible.
func seedFourQueueConfigurations(t *testing.T) {
	t.Helper()

	if err := GetDB().Create(&models.SystemSettings{}).Error; err != nil {
		t.Fatalf("failed to create system settings: %v", err)
	}

	priority := 11
	for _, queueName := range models.GetQueueNames() {
		if queueName == models.SystemCleanUpQueue {
			continue
		}

		err := GetDB().Create(&models.TaskQueueConfiguration{
			Name:             queueName,
			Priority:         priority,
			SystemSettingsId: 1,
		}).Error
		if err != nil {
			t.Fatalf("failed to seed queue configuration %s: %v", queueName, err)
		}

		priority++
	}
}

// An install upgraded across the release that added system_clean_up has four
// persisted configurations. The read used to substitute defaults only when the list
// was EMPTY, so it returned four -- the settings form then submitted four and
// UpsertSystemSettingsCommand.Validate, which requires one per queue name, rejected
// every save with a 400.
func TestGetSystemSettingsFillsInAMissingQueueConfiguration(t *testing.T) {
	defer TruncateTestDb()
	seedFourQueueConfigurations(t)

	settings, err := NewSystemSettingsRepository(nil).GetSystemSettings()
	if err != nil {
		t.Fatalf("failed to read system settings: %v", err)
	}

	queueNames := models.GetQueueNames()
	if len(settings.TaskQueueConfigurations) != len(queueNames) {
		t.Fatalf("got %d queue configurations, expected %d", len(settings.TaskQueueConfigurations), len(queueNames))
	}

	byName := map[models.QueueName]int{}
	for _, configuration := range settings.TaskQueueConfigurations {
		byName[configuration.Name] = configuration.Priority
	}

	for _, queueName := range queueNames {
		if _, ok := byName[queueName]; !ok {
			t.Fatalf("queue configuration for %s is missing", queueName)
		}
	}

	// The persisted priorities must survive the merge -- filling the gap must not
	// reset the queues the admin already configured.
	expected := 11
	for _, queueName := range queueNames {
		if queueName == models.SystemCleanUpQueue {
			continue
		}

		if byName[queueName] != expected {
			t.Errorf("%s priority = %d, expected the persisted %d", queueName, byName[queueName], expected)
		}

		expected++
	}

	defaultPriority := models.GetDefaultSystemCleanupQueueConfiguration().Priority
	if byName[models.SystemCleanUpQueue] != defaultPriority {
		t.Errorf("%s priority = %d, expected the default %d", models.SystemCleanUpQueue, byName[models.SystemCleanUpQueue], defaultPriority)
	}
}

// The read-side fill alone leaves the database broken: the update path only ever
// UPDATEd by name, so the submitted priority for the missing queue matched no row
// and was silently discarded on every save, forever.
func TestUpdateSystemSettingsInsertsAMissingQueueConfiguration(t *testing.T) {
	defer TruncateTestDb()
	seedFourQueueConfigurations(t)

	command := buildSettingsCommand()
	for i := range command.TaskQueueConfigurations {
		if command.TaskQueueConfigurations[i].Name == models.SystemCleanUpQueue {
			command.TaskQueueConfigurations[i].Priority = 7
		}
	}

	if _, err := NewSystemSettingsRepository(nil).UpdateSystemSettings(command); err != nil {
		t.Fatalf("failed to update system settings: %v", err)
	}

	var persisted []models.TaskQueueConfiguration
	if err := GetDB().Find(&persisted).Error; err != nil {
		t.Fatalf("failed to read queue configurations: %v", err)
	}

	queueNames := models.GetQueueNames()
	if len(persisted) != len(queueNames) {
		t.Fatalf("got %d persisted queue configurations, expected %d", len(persisted), len(queueNames))
	}

	byName := map[models.QueueName]models.TaskQueueConfiguration{}
	for _, configuration := range persisted {
		byName[configuration.Name] = configuration
	}

	inserted, ok := byName[models.SystemCleanUpQueue]
	if !ok {
		t.Fatalf("no row was inserted for %s", models.SystemCleanUpQueue)
	}

	if inserted.Priority != 7 {
		t.Errorf("%s priority = %d, expected the submitted 7", models.SystemCleanUpQueue, inserted.Priority)
	}

	if inserted.SystemSettingsId != 1 {
		t.Errorf("%s systemSettingsId = %d, expected 1", models.SystemCleanUpQueue, inserted.SystemSettingsId)
	}

	// The queues that already had rows still take the submitted priority, so the
	// insert did not come at the cost of the update it sits beside.
	for _, queueName := range queueNames {
		if queueName == models.SystemCleanUpQueue {
			continue
		}

		if byName[queueName].Priority != 1 {
			t.Errorf("%s priority = %d, expected the submitted 1", queueName, byName[queueName].Priority)
		}
	}
}

// A rolled-back transaction has to reach the caller as an error. It used to return
// (zero settings, nil), so the handler answered 200 with an empty body and the
// admin's change vanished without a signal -- and, worse, every caller's `err`
// check was silently disarmed, including the one in the test above.
//
// The failure is forced through the insert this feature added: an unknown queue
// name reaches tx.Create, where models.QueueName.Value() refuses it.
func TestUpdateSystemSettingsReturnsTheTransactionError(t *testing.T) {
	defer TruncateTestDb()
	seedFourQueueConfigurations(t)

	command := buildSettingsCommand()
	command.TaskConcurrency = 9
	for i := range command.TaskQueueConfigurations {
		if command.TaskQueueConfigurations[i].Name == models.SystemCleanUpQueue {
			command.TaskQueueConfigurations[i].Name = models.QueueName("not_a_real_queue")
		}
	}

	_, err := NewSystemSettingsRepository(nil).UpdateSystemSettings(command)
	if err == nil {
		t.Fatal("expected the failed transaction to be reported to the caller")
	}

	// And the rollback really rolled back, so the error is not merely cosmetic.
	settings, readErr := NewSystemSettingsRepository(nil).GetSystemSettings()
	if readErr != nil {
		t.Fatalf("failed to read system settings: %v", readErr)
	}

	if settings.TaskConcurrency == 9 {
		t.Error("the scalar update survived a rolled-back transaction")
	}

	var persistedCount int64
	if err := GetDB().Model(&models.TaskQueueConfiguration{}).Count(&persistedCount).Error; err != nil {
		t.Fatalf("failed to count queue configurations: %v", err)
	}

	if persistedCount != 4 {
		t.Errorf("got %d persisted queue configurations, expected the original 4", persistedCount)
	}

	// The row count alone would miss the statements this feature actually added:
	// the per-name loop updates all four existing queues to the submitted priority
	// BEFORE it reaches the insert that fails, so an un-rolled-back transaction
	// leaves four rows that all read 1.
	var persisted []models.TaskQueueConfiguration
	if err := GetDB().Find(&persisted).Error; err != nil {
		t.Fatalf("failed to read queue configurations: %v", err)
	}

	priorities := make(map[models.QueueName]int, len(persisted))
	for _, configuration := range persisted {
		priorities[configuration.Name] = configuration.Priority
	}

	expected := 11
	for _, queueName := range models.GetQueueNames() {
		if queueName == models.SystemCleanUpQueue {
			continue
		}

		priority, ok := priorities[queueName]
		if !ok {
			t.Errorf("queue configuration for %s was not persisted", queueName)
		} else if priority != expected {
			t.Errorf("%s priority = %d, expected the seeded %d", queueName, priority, expected)
		}

		expected++
	}
}
