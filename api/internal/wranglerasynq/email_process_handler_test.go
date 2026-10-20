package wranglerasynq

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

func TestEnsureOcrCopy(t *testing.T) {
	cases := []struct {
		name     string
		original []byte
		existing []byte
		outside  bool
		wantErr  bool
		wantCopy []byte
	}{
		{name: "missing copy is made from the original", original: pollJpeg, wantCopy: pollJpeg},
		{
			// A payload queued before the poll stopped converting, or an attachment
			// a sibling group's task converted first.
			name:     "an existing copy is left alone",
			original: pollJpeg,
			existing: []byte("already converted"),
			wantCopy: []byte("already converted"),
		},
		{name: "a path outside temp/ is refused", original: pollJpeg, outside: true, wantErr: true},
		{name: "a file that will not convert leaves nothing behind", original: corruptPdf, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(repositories.TruncateTestDb)
			directory := containedTempDir(t)

			path := filepath.Join(directory, "image-attachment")
			if tc.outside {
				path = filepath.Join(filepath.Dir(directory), "image-attachment")
			}
			if tc.existing != nil {
				if err := os.WriteFile(path, tc.existing, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			err := ensureOcrCopy(path, tc.original)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ensureOcrCopy() = %v, want an error: %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
					t.Errorf("a copy was written although it failed (stat: %v)", statErr)
				}
				entries, _ := os.ReadDir(filepath.Dir(path))
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".tmp") {
						t.Errorf("left a partial file behind: %s", entry.Name())
					}
				}
				return
			}
			if got := readFile(t, path); !bytes.Equal(got, tc.wantCopy) {
				t.Errorf("copy = %q, want %q", got, tc.wantCopy)
			}
		})
	}
}

func TestProcessEmailTask_ConversionFailureKeepsTheImageAndIsRecorded(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, _ := fixture.addSystemEmail("receipts@example.org")
	settings := fixture.addGroup(systemEmail, groupOptions{})

	original, err := repositories.NewFileRepository(nil).WriteTempFile(corruptPdf)
	if err != nil {
		t.Fatal(err)
	}
	payload := EmailProcessTaskPayload{
		GroupSettingsId: settings.ID,
		TempFilePath:    original,
		ImageForOcrPath: filepath.Join(filepath.Dir(original), "image-"+filepath.Base(original)),
		Metadata:        structs.EmailMetadata{Subject: "Lunch", FromEmail: "bob@example.com"},
		Attachment:      structs.Attachment{Filename: "receipt.pdf", FileType: "application/pdf"},
	}

	err = processEmailTask("task-1", payload)

	// Converting the same bytes again would fail the same way, so the task is
	// archived at once rather than retried.
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("processEmailTask() = %v, want it to skip the retries", err)
	}

	emailReads := fixture.systemTasks(models.EMAIL_READ)
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(emailReads) != 1 || len(uploads) != 1 {
		t.Fatalf("EMAIL_READ = %d, EMAIL_UPLOAD = %d, want one of each", len(emailReads), len(uploads))
	}
	emailRead, upload := emailReads[0], uploads[0]
	if emailRead.Status != models.SYSTEM_TASK_FAILED || emailRead.AsynqTaskId != "task-1" ||
		!strings.Contains(emailRead.ResultDescription, `"receipt.pdf" could not be converted`) {
		t.Errorf("EMAIL_READ = %+v", emailRead)
	}
	// The group's feed lists the upload, and the task id on it is what offers the
	// source file and the rerun.
	if upload.Status != models.SYSTEM_TASK_FAILED || upload.AsynqTaskId != "task-1" ||
		upload.GroupId == nil || *upload.GroupId != settings.GroupId ||
		upload.AssociatedSystemTaskId == nil || *upload.AssociatedSystemTaskId != emailRead.ID {
		t.Errorf("EMAIL_UPLOAD = %+v", upload)
	}

	if !bytes.Equal(readFile(t, original), corruptPdf) {
		t.Error("the original was not kept")
	}
	if _, err := os.Stat(payload.ImageForOcrPath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("an OCR copy exists after a failed conversion (stat: %v)", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if !RerunSourceFilesPresent(models.EMAIL_UPLOAD, payloadBytes) {
		t.Error("a rerun is refused, but the original it needs is still there")
	}
}
