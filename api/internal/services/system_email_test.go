package services

import (
	"errors"
	"testing"

	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/emailclient"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

const connectivityTestKey = "connectivity-test-encryption-key"

// stubEmailConnection replaces the IMAP login with one that records the
// credentials it was given and returns result.
func stubEmailConnection(t *testing.T, result error) *[]emailclient.Credentials {
	t.Helper()
	calls := make([]emailclient.Credentials, 0)
	original := checkEmailConnection
	checkEmailConnection = func(credentials emailclient.Credentials) error {
		calls = append(calls, credentials)
		return result
	}
	t.Cleanup(func() { checkEmailConnection = original })
	return &calls
}

func createStoredSystemEmail(t *testing.T) models.SystemEmail {
	t.Helper()
	t.Setenv("ENCRYPTION_KEY", connectivityTestKey)
	password, err := utils.EncryptAndEncodeToBase64(connectivityTestKey, "stored secret")
	if err != nil {
		t.Fatal(err)
	}

	systemEmail := models.SystemEmail{Host: "imap.example.org", Port: "143", Username: "receipts@example.org", Password: password, UseStartTLS: true}
	if err := repositories.GetDB().Create(&systemEmail).Error; err != nil {
		t.Fatal(err)
	}
	return systemEmail
}

func storedConnectivityTasks(t *testing.T) []models.SystemTask {
	t.Helper()
	var tasks []models.SystemTask
	if err := repositories.GetDB().Where("type = ?", models.SYSTEM_EMAIL_CONNECTIVITY_CHECK).Find(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	return tasks
}

func TestCheckEmailConnectivity_StoredSystemEmailSucceeds(t *testing.T) {
	defer repositories.TruncateTestDb()
	systemEmail := createStoredSystemEmail(t)
	calls := stubEmailConnection(t, nil)

	task, err := NewSystemEmailService(nil).CheckEmailConnectivity(commands.CheckEmailConnectivityCommand{ID: systemEmail.ID}, 7)
	if err != nil {
		t.Fatal(err)
	}

	want := emailclient.Credentials{Host: "imap.example.org", Port: "143", Username: "receipts@example.org", Password: "stored secret", UseStartTLS: true}
	if len(*calls) != 1 || (*calls)[0] != want {
		t.Errorf("credentials = %+v, want the stored ones with the password decrypted", *calls)
	}
	if task.Status != models.SYSTEM_TASK_SUCCEEDED || task.ResultDescription != "Connection successful" {
		t.Errorf("task = %+v", task)
	}

	stored := storedConnectivityTasks(t)
	if len(stored) != 1 || stored[0].AssociatedEntityId != systemEmail.ID || stored[0].RanByUserId == nil || *stored[0].RanByUserId != 7 || stored[0].EndedAt == nil {
		t.Errorf("stored tasks = %+v", stored)
	}
}

func TestCheckEmailConnectivity_FailureReportsTheErrorItself(t *testing.T) {
	defer repositories.TruncateTestDb()
	systemEmail := createStoredSystemEmail(t)
	stubEmailConnection(t, errors.New("NO [AUTHENTICATIONFAILED] Invalid credentials"))

	task, err := NewSystemEmailService(nil).CheckEmailConnectivity(commands.CheckEmailConnectivityCommand{ID: systemEmail.ID}, 1)
	if err != nil {
		t.Fatalf("error = %v; a failed check is reported on the task, not as an error", err)
	}

	// The Python client's error was scraped from the last line of a log file
	// shared with every poll, so it could report another mailbox's failure.
	if task.Status != models.SYSTEM_TASK_FAILED || task.ResultDescription != "NO [AUTHENTICATIONFAILED] Invalid credentials" {
		t.Errorf("task = %+v", task)
	}
	if stored := storedConnectivityTasks(t); len(stored) != 1 || stored[0].Status != models.SYSTEM_TASK_FAILED {
		t.Errorf("stored tasks = %+v", stored)
	}
}

func TestCheckEmailConnectivity_UnsavedSystemEmailIsNotRecorded(t *testing.T) {
	defer repositories.TruncateTestDb()
	calls := stubEmailConnection(t, errors.New("dial tcp: connection refused"))

	command := commands.CheckEmailConnectivityCommand{}
	command.Host = "imap.example.org"
	command.Port = "993"
	command.Username = "new@example.org"
	command.Password = "typed secret"

	task, err := NewSystemEmailService(nil).CheckEmailConnectivity(command, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(*calls) != 1 || (*calls)[0].Password != "typed secret" || (*calls)[0].UseStartTLS {
		t.Errorf("credentials = %+v, want the submitted ones", *calls)
	}
	if task.Status != models.SYSTEM_TASK_FAILED || task.ResultDescription != "dial tcp: connection refused" {
		t.Errorf("task = %+v", task)
	}
	if stored := storedConnectivityTasks(t); len(stored) != 0 {
		t.Errorf("an unsaved system email's check was stored: %+v", stored)
	}
}
