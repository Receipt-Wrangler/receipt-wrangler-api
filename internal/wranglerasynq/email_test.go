package wranglerasynq

import (
	miniredis2 "github.com/alicebob/miniredis/v2"
	"os"
	"receipt-wrangler/api/internal/repositories"
	"testing"
)

var miniredisInstance *miniredis2.Miniredis

func initEmailProcessHandlerTest() {
	instance, err := miniredis2.Run()
	if err != nil {
		panic(err)
	}

	miniredisInstance = instance

	os.Setenv("REDIS_HOST", miniredisInstance.Host())
	os.Setenv("REDIS_PORT", miniredisInstance.Port())

	err = repositories.ConnectToRedis()
	if err != nil {
		panic(err)
	}
}

func TestEnqueueEmailProcessTasksShouldEnqueueTasks(t *testing.T) {
	/*
		TODO: This passes locally but not in CI. Fix
		// Setup and teardown
		utils.MakeDirectory(env.GetBasePath() + "/temp")
		defer teardownEmailProcessHandlerTest()
		initEmailProcessHandlerTest()
		fileRepository := repositories.NewFileRepository(nil)

		// Test data setup
		attachments := []structs.Attachment{
			{
				Filename: "test.jpg",
				FileType: "image/jpeg",
				Size:     0,
			},
			{
				Filename: "test2.jpg",
				FileType: "image/jpeg",
				Size:     0,
			},
		}

		metadata := structs.EmailMetadata{
			Date:             time.Time{},
			Subject:          "Test",
			To:               "test@test.com",
			FromName:         "Test Man",
			FromEmail:        "test@test.com",
			Attachments:      attachments,
			GroupSettingsIds: []uint{1, 2},
		}

		// Create temporary files for testing
		for _, attachment := range attachments {
			testJpgBytes, err := fileRepository.GetTestJpgBytes()
			if err != nil {
				t.Fatalf("Failed to get test jpg bytes: %v", err)
			}

			filePath := buildTempEmailFilePath(attachment.Filename)
			err = utils.WriteFile(filePath, testJpgBytes)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Clean up files after test
			defer os.Remove(filePath)
			defer os.Remove(buildTempEmailOcrFilePath(attachment.Filename))
		}

		// Enqueue tasks
		err := enqueueEmailProcessTasks([]structs.EmailMetadata{metadata})
		if err != nil {
			t.Fatalf("Failed to enqueue tasks: %v", err)
		}

		// Verify tasks were created correctly
		inspector, err := GetAsynqInspector()
		if err != nil {
			t.Fatalf("Failed to get inspector: %v", err)
		}

		taskInfos, err := inspector.ListPendingTasks(string(models.EmailReceiptProcessingQueue))
		if err != nil {
			t.Fatalf("Failed to list pending tasks: %v", err)
		}

		// Check number of tasks
		if len(taskInfos) != 4 {
			t.Errorf("Expected 4 tasks, got %d", len(taskInfos))
		}

		// Verify each task's payload
		for index, taskInfo := range taskInfos {
			var payload EmailProcessTaskPayload
			err := json.Unmarshal(taskInfo.Payload, &payload)
			if err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			expectedAttachment := attachments[index/2]
			expectedPayload := EmailProcessTaskPayload{
				GroupSettingsId: uint((index % 2) + 1),
				ImageForOcrPath: buildTempEmailOcrFilePath(expectedAttachment.Filename),
				TempFilePath:    buildTempEmailFilePath(expectedAttachment.Filename),
				Metadata:        metadata,
				Attachment:      expectedAttachment,
			}

			if !reflect.DeepEqual(payload, expectedPayload) {
				t.Errorf("Task %d payload mismatch:\nexpected: %+v\ngot: %+v",
					index, expectedPayload, payload)
			}
		}*/
}

func teardownEmailProcessHandlerTest() {
	miniredisInstance.Close()
	repositories.ShutdownAsynqClient()
	teardown()
}
