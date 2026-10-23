package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func createGroupWithSettings(t *testing.T) models.GroupSettings {
	t.Helper()
	CreateTestGroup()

	repository := NewGroupSettingsRepository(nil)
	created, err := repository.CreateGroupSettings(models.GroupSettings{GroupId: 1})
	if err != nil {
		t.Fatal(err)
	}

	// Reload so we see the DB-applied column default rather than the struct.
	settings, err := repository.GetGroupSettingsById(utils.UintToString(created.ID))
	if err != nil {
		t.Fatal(err)
	}
	return settings
}

func TestCreateGroupSettingsRequiresASignedSenderByDefault(t *testing.T) {
	defer TruncateTestDb()

	settings := createGroupWithSettings(t)

	if !settings.EmailRequireSignedSender {
		utils.PrintTestError(t, settings.EmailRequireSignedSender, true)
	}
}

func TestUpdateGroupSettingsEmailRequireSignedSender(t *testing.T) {
	defer TruncateTestDb()
	createGroupWithSettings(t)
	repository := NewGroupSettingsRepository(nil)

	steps := []struct {
		name string
		sent *bool
		want bool
	}{
		// A client that predates the setting leaves it out of the body.
		{"omitted leaves the default", nil, true},
		{"explicit false is persisted", boolPointer(false), false},
		{"omitted leaves the stored false", nil, false},
		{"explicit true is persisted", boolPointer(true), true},
	}

	for _, step := range steps {
		_, err := repository.UpdateGroupSettings("1", commands.UpdateGroupSettingsCommand{EmailRequireSignedSender: step.sent})
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}

		settings, err := repository.GetAllGroupSettings("group_id = ?", 1)
		if err != nil || len(settings) != 1 {
			t.Fatalf("%s: %v %v", step.name, settings, err)
		}
		if settings[0].EmailRequireSignedSender != step.want {
			t.Errorf("%s: EmailRequireSignedSender = %v, want %v", step.name, settings[0].EmailRequireSignedSender, step.want)
		}
	}
}

func TestGroupSettingsSignedSenderBackfillsExistingRows(t *testing.T) {
	defer TruncateTestDb()
	db := GetDB()
	migrator := db.Migrator()

	// Recreate the pre-upgrade schema: no column, one existing row.
	if err := migrator.DropColumn(&models.GroupSettings{}, "email_require_signed_sender"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.AutoMigrate(&models.GroupSettings{}); err != nil {
			t.Fatal(err)
		}
	})

	CreateTestGroup()
	if err := db.Exec("INSERT INTO group_settings (group_id) VALUES (?)", 1).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.GroupSettings{}); err != nil {
		t.Fatal(err)
	}

	var settings []models.GroupSettings
	if err := db.Find(&settings).Error; err != nil || len(settings) != 1 {
		t.Fatalf("%v %v", settings, err)
	}
	if !settings[0].EmailRequireSignedSender {
		utils.PrintTestError(t, settings[0].EmailRequireSignedSender, "true: an upgraded install starts with the check on")
	}
}

func boolPointer(value bool) *bool {
	return &value
}
