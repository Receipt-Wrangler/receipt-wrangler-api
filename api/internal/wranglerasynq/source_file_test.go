package wranglerasynq

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

func TestResolveActivityFlags(t *testing.T) {
	directory := containedTempDir(t)
	present := writeTempFile(t, directory, "present.jpg", 0)
	presentOcr := writeTempFile(t, directory, "image-present.jpg", 0)
	missing := filepath.Join(directory, "gone.jpg")

	tests := []struct {
		name                   string
		taskType               models.SystemTaskType
		state                  asynq.TaskState
		payload                any
		expectedCanBeRestarted bool
		expectedHasSourceFile  bool
	}{
		{
			name:                   "archived quick scan with its upload is rerunnable and downloadable",
			taskType:               models.QUICK_SCAN,
			state:                  asynq.TaskStateArchived,
			payload:                QuickScanTaskPayload{TempPath: present},
			expectedCanBeRestarted: true,
			expectedHasSourceFile:  true,
		},
		{
			// Once the sweeper reclaims the file, a rerun cannot work and the
			// button must stop being offered.
			name:                   "archived quick scan without its upload is neither",
			taskType:               models.QUICK_SCAN,
			state:                  asynq.TaskStateArchived,
			payload:                QuickScanTaskPayload{TempPath: missing},
			expectedCanBeRestarted: false,
			expectedHasSourceFile:  false,
		},
		{
			// The point of not gating HasSourceFile on Archived: asynq backs its
			// retries off exponentially, so this state lasts minutes.
			name:                   "retrying quick scan offers the file but not a rerun",
			taskType:               models.QUICK_SCAN,
			state:                  asynq.TaskStateRetry,
			payload:                QuickScanTaskPayload{TempPath: present},
			expectedCanBeRestarted: false,
			expectedHasSourceFile:  true,
		},
		{
			name:                   "pending quick scan offers the file but not a rerun",
			taskType:               models.QUICK_SCAN,
			state:                  asynq.TaskStatePending,
			payload:                QuickScanTaskPayload{TempPath: present},
			expectedCanBeRestarted: false,
			expectedHasSourceFile:  true,
		},
		{
			// A body-only email expects no upload at all, so it stays rerunnable.
			// This is why "expects a file" is tracked apart from "has a file".
			name:                   "archived body-only email is rerunnable with no source file",
			taskType:               models.EMAIL_UPLOAD,
			state:                  asynq.TaskStateArchived,
			payload:                EmailProcessTaskPayload{},
			expectedCanBeRestarted: true,
			expectedHasSourceFile:  false,
		},
		{
			name:                   "archived email with both files is rerunnable and downloadable",
			taskType:               models.EMAIL_UPLOAD,
			state:                  asynq.TaskStateArchived,
			payload:                EmailProcessTaskPayload{TempFilePath: present, ImageForOcrPath: presentOcr},
			expectedCanBeRestarted: true,
			expectedHasSourceFile:  true,
		},
		{
			// The email queue retains its completed tasks so the temp sweep can
			// observe them and release their files (see completedTaskRetention),
			// which means a succeeded upload is now visible here with both files
			// still on disk. Neither control belongs on it: the receipt already
			// carries the image, and a rerun would duplicate it.
			name:                   "completed email offers neither, though both files survive",
			taskType:               models.EMAIL_UPLOAD,
			state:                  asynq.TaskStateCompleted,
			payload:                EmailProcessTaskPayload{TempFilePath: present, ImageForOcrPath: presentOcr},
			expectedCanBeRestarted: false,
			expectedHasSourceFile:  false,
		},
		{
			// The task recreates the OCR copy from the original, so a missing copy,
			// or one a failed conversion never made, does not stop a rerun.
			name:                   "email missing only its OCR copy is rerunnable and downloadable",
			taskType:               models.EMAIL_UPLOAD,
			state:                  asynq.TaskStateArchived,
			payload:                EmailProcessTaskPayload{TempFilePath: present, ImageForOcrPath: missing},
			expectedCanBeRestarted: true,
			expectedHasSourceFile:  true,
		},
		{
			// The original is the one file nothing can recreate.
			name:                   "email missing its original is neither",
			taskType:               models.EMAIL_UPLOAD,
			state:                  asynq.TaskStateArchived,
			payload:                EmailProcessTaskPayload{TempFilePath: missing, ImageForOcrPath: presentOcr},
			expectedCanBeRestarted: false,
			expectedHasSourceFile:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lookup := lookupReturning(t, test.state, test.payload)

			flags, err := resolveActivityFlags(lookup, test.taskType, "task-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if flags.canBeRestarted != test.expectedCanBeRestarted {
				t.Errorf("canBeRestarted = %v, want %v", flags.canBeRestarted, test.expectedCanBeRestarted)
			}
			if flags.hasSourceFile != test.expectedHasSourceFile {
				t.Errorf("hasSourceFile = %v, want %v", flags.hasSourceFile, test.expectedHasSourceFile)
			}
		})
	}
}

// RECEIPT_UPLOADED / RECEIPT_UPDATED rows carry no asynq task, and GetTaskInfo
// rejects an empty id — so it must never be called for one.
func TestResolveActivityFlags_SkipsLookupForRowsWithNoTask(t *testing.T) {
	called := false
	lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
		called = true
		return nil, nil
	}

	flags, err := resolveActivityFlags(lookup, models.QUICK_SCAN, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no lookup for an empty asynq task id")
	}
	if flags.canBeRestarted || flags.hasSourceFile {
		t.Error("expected no flags")
	}
}

func TestResolveActivityFlags_UnsupportedTaskTypeIsNotAnError(t *testing.T) {
	lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
		t.Fatal("expected no lookup for an unsupported task type")
		return nil, nil
	}

	if _, err := resolveActivityFlags(lookup, models.RECEIPT_UPLOADED, "task-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A task that aged out of Redis is an ordinary end state, not a failure.
func TestResolveActivityFlags_MissingTaskYieldsNoFlags(t *testing.T) {
	for _, sentinel := range []error{asynq.ErrTaskNotFound, asynq.ErrQueueNotFound} {
		lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
			return nil, fmt.Errorf("asynq: %w", sentinel)
		}

		flags, err := resolveActivityFlags(lookup, models.QUICK_SCAN, "task-1")
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", sentinel, err)
		}
		if flags.canBeRestarted || flags.hasSourceFile {
			t.Errorf("expected no flags for %v", sentinel)
		}
	}
}

// Any other lookup failure propagates, so the caller can stop rather than pay a
// dial timeout per remaining row.
func TestResolveActivityFlags_PropagatesOtherLookupErrors(t *testing.T) {
	lookupErr := errors.New("redis unreachable")
	lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
		return nil, lookupErr
	}

	if _, err := resolveActivityFlags(lookup, models.QUICK_SCAN, "task-1"); !errors.Is(err, lookupErr) {
		t.Fatalf("expected the lookup error to propagate, got: %v", err)
	}
}

// An unreadable payload is this row's problem, not the request's.
func TestResolveActivityFlags_MalformedPayloadYieldsNoFlags(t *testing.T) {
	lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
		return &asynq.TaskInfo{State: asynq.TaskStateArchived, Payload: []byte("{not json")}, nil
	}

	flags, err := resolveActivityFlags(lookup, models.QUICK_SCAN, "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.canBeRestarted || flags.hasSourceFile {
		t.Error("expected no flags")
	}
}

// A primary and fallback processing setting produce two rows sharing one asynq
// task id, and the system-task table shows parent and child rows, so the same id
// repeats several times per page.
func TestMemoizeTaskInfoLookup_LooksEachTaskUpOnce(t *testing.T) {
	calls := 0
	lookup := memoizeTaskInfoLookup(func(queue string, id string) (*asynq.TaskInfo, error) {
		calls++
		return &asynq.TaskInfo{State: asynq.TaskStateArchived}, nil
	})

	for i := 0; i < 3; i++ {
		if _, err := lookup("quick_scan", "task-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if _, err := lookup("quick_scan", "task-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls != 2 {
		t.Errorf("expected 2 lookups, got %d", calls)
	}
}

func TestRerunSourceFilesPresent(t *testing.T) {
	directory := containedTempDir(t)
	present := writeTempFile(t, directory, "present.jpg", 0)
	missing := filepath.Join(directory, "gone.jpg")

	tests := []struct {
		name     string
		taskType models.SystemTaskType
		payload  any
		expected bool
	}{
		{"quick scan with its upload", models.QUICK_SCAN, QuickScanTaskPayload{TempPath: present}, true},
		{"quick scan without its upload", models.QUICK_SCAN, QuickScanTaskPayload{TempPath: missing}, false},
		{"body-only email needs nothing", models.EMAIL_UPLOAD, EmailProcessTaskPayload{}, true},
		{
			"email missing only its OCR copy, which the task recreates",
			models.EMAIL_UPLOAD,
			EmailProcessTaskPayload{TempFilePath: present, ImageForOcrPath: missing},
			true,
		},
		{
			"email missing its original",
			models.EMAIL_UPLOAD,
			EmailProcessTaskPayload{TempFilePath: missing, ImageForOcrPath: present},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(test.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			if got := RerunSourceFilesPresent(test.taskType, payloadBytes); got != test.expected {
				t.Errorf("RerunSourceFilesPresent() = %v, want %v", got, test.expected)
			}
		})
	}
}

// The obvious "top level means no parent" rule is wrong here: every EMAIL_UPLOAD
// task is chained under its EMAIL_READ parent, so that rule would strip the flag
// from exactly the rows this feature exists for.
func TestSetSystemTaskHasSourceFile_HydratesEmailUploadWithAParent(t *testing.T) {
	directory := containedTempDir(t)
	present := writeTempFile(t, directory, "attachment.pdf", 0)

	parentId := uint(42)
	groupId := uint(7)
	systemTasks := []models.SystemTask{
		{
			Type:                   models.EMAIL_UPLOAD,
			AsynqTaskId:            "email-1",
			GroupId:                &groupId,
			AssociatedSystemTaskId: &parentId,
		},
	}

	lookup := lookupReturning(t, asynq.TaskStateArchived, EmailProcessTaskPayload{
		TempFilePath:    present,
		ImageForOcrPath: present,
	})

	hydrateSystemTaskSourceFiles(systemTasks, lookup, allowAllGroups)

	if !systemTasks[0].HasSourceFile {
		t.Error("an EMAIL_UPLOAD row with a parent must still be hydrated")
	}
}

// The system-task table is app-scoped, so it lists groups the caller may not
// belong to. A flag there would advertise a button the endpoints refuse.
func TestSetSystemTaskHasSourceFile_SkipsGroupsTheCallerCannotRead(t *testing.T) {
	directory := containedTempDir(t)
	present := writeTempFile(t, directory, "scan.jpg", 0)

	readable := uint(1)
	unreadable := uint(2)
	systemTasks := []models.SystemTask{
		{Type: models.QUICK_SCAN, AsynqTaskId: "task-1", GroupId: &readable},
		{Type: models.QUICK_SCAN, AsynqTaskId: "task-2", GroupId: &unreadable},
		{Type: models.QUICK_SCAN, AsynqTaskId: "task-3", GroupId: nil},
	}

	lookup := lookupReturning(t, asynq.TaskStateArchived, QuickScanTaskPayload{TempPath: present})

	hydrateSystemTaskSourceFiles(systemTasks, lookup, func(groupId uint) bool { return groupId == readable })

	if !systemTasks[0].HasSourceFile {
		t.Error("expected the readable group's row to be hydrated")
	}
	if systemTasks[1].HasSourceFile {
		t.Error("expected the unreadable group's row to be skipped")
	}
	if systemTasks[2].HasSourceFile {
		t.Error("expected a row with no group to be skipped")
	}
}

func TestSetActivityFlags_AppliesBothFlagsPerRow(t *testing.T) {
	directory := containedTempDir(t)
	present := writeTempFile(t, directory, "scan.jpg", 0)
	missing := filepath.Join(directory, "gone.jpg")

	activities := []structs.Activity{
		{Type: models.QUICK_SCAN, AsynqTaskId: "task-present"},
		{Type: models.QUICK_SCAN, AsynqTaskId: "task-missing"},
	}

	lookup := func(queue string, id string) (*asynq.TaskInfo, error) {
		tempPath := present
		if id == "task-missing" {
			tempPath = missing
		}

		payload, err := json.Marshal(QuickScanTaskPayload{TempPath: tempPath})
		if err != nil {
			return nil, err
		}

		return &asynq.TaskInfo{State: asynq.TaskStateArchived, Payload: payload}, nil
	}

	applyActivityFlags(&activities, lookup)

	if !activities[0].CanBeRestarted || !activities[0].HasSourceFile {
		t.Error("expected the activity with its upload to carry both flags")
	}
	if activities[1].CanBeRestarted || activities[1].HasSourceFile {
		t.Error("expected the activity without its upload to carry neither flag")
	}
}

// --- helpers ---

func allowAllGroups(groupId uint) bool { return true }

func lookupReturning(t *testing.T, state asynq.TaskState, payload any) taskInfoLookup {
	t.Helper()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	return func(queue string, id string) (*asynq.TaskInfo, error) {
		return &asynq.TaskInfo{ID: id, State: state, Payload: payloadBytes}, nil
	}
}

// Exhaustive over every asynq state, so a state added upstream is a deliberate
// decision rather than a silent default.
//
// These two predicates are what keeps asynq.Retention on the email queue from
// changing anything a user sees: RunTask does not refuse a task by state, and a
// retained completed task still has its files on disk.
func TestTaskStatePredicates(t *testing.T) {
	tests := []struct {
		state                   asynq.TaskState
		expectedRerunnable      bool
		expectedOffersSourceFil bool
	}{
		{asynq.TaskStateActive, false, true},
		{asynq.TaskStatePending, false, true},
		{asynq.TaskStateScheduled, false, true},
		{asynq.TaskStateRetry, false, true},
		// The state a failure lands in once asynq exhausts its retries — the one
		// a rerun is for.
		{asynq.TaskStateArchived, true, true},
		// Retained for the temp sweep. It succeeded: its receipt carries the
		// image, and re-running it would make a second one.
		{asynq.TaskStateCompleted, false, false},
		{asynq.TaskStateAggregating, false, true},
	}

	for _, test := range tests {
		t.Run(test.state.String(), func(t *testing.T) {
			if got := rerunnableState(test.state); got != test.expectedRerunnable {
				t.Errorf("rerunnableState(%v) = %v, want %v", test.state, got, test.expectedRerunnable)
			}
			if got := CanRerunTask(&asynq.TaskInfo{State: test.state}); got != test.expectedRerunnable {
				t.Errorf("CanRerunTask(%v) = %v, want %v", test.state, got, test.expectedRerunnable)
			}
			if got := offersSourceFile(test.state); got != test.expectedOffersSourceFil {
				t.Errorf("offersSourceFile(%v) = %v, want %v", test.state, got, test.expectedOffersSourceFil)
			}
		})
	}
}

// The zero TaskState is what an unset struct carries; it must not read as
// rerunnable, and neither must a nil lookup result.
func TestCanRerunTask_RefusesNilAndZeroState(t *testing.T) {
	if CanRerunTask(nil) {
		t.Error("CanRerunTask(nil) = true, want false")
	}
	if CanRerunTask(&asynq.TaskInfo{}) {
		t.Error("CanRerunTask(zero state) = true, want false")
	}
}

// containedTempDir points BASE_PATH at a fresh directory and returns its temp/
// subdirectory, which is the only place sourcePathUsable accepts a path from.
//
// A bare t.TempDir() is NOT inside temp/, so every fixture written there would read as
// missing once the flag path started asserting containment. Rooting the fixtures properly
// exercises the real check rather than injecting a stub for it.
func containedTempDir(t *testing.T) string {
	t.Helper()

	base := t.TempDir()
	directory := filepath.Join(base, "temp")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", directory, err)
	}
	t.Setenv("BASE_PATH", base)

	return directory
}

// A payload naming a path outside temp/ must report no source file, matching what
// ResolveSystemTaskSourceFile does at the endpoint. Otherwise the client renders a preview
// control whose request can only fail, and the stat doubles as an existence oracle for
// arbitrary server paths.
func TestResolveActivityFlags_PathOutsideTempIsNotUsable(t *testing.T) {
	outside := t.TempDir()
	escaped := filepath.Join(outside, "escaped.jpg")
	if err := os.WriteFile(escaped, []byte("contents"), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", escaped, err)
	}

	// BASE_PATH is set to a DIFFERENT root, so temp/ does not contain the file above.
	containedTempDir(t)

	lookup := lookupReturning(t, asynq.TaskStateArchived, QuickScanTaskPayload{TempPath: escaped})

	flags, err := resolveActivityFlags(lookup, models.QUICK_SCAN, "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.hasSourceFile {
		t.Error("hasSourceFile = true for a path outside temp/, want false")
	}
	if flags.canBeRestarted {
		t.Error("canBeRestarted = true for a path outside temp/, want false")
	}
}
