package wranglerasynq

import (
	"encoding/json"
	miniredis2 "github.com/alicebob/miniredis/v2"
	"os"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"reflect"
	"testing"
	"time"
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
	defer teardownEmailProcessHandlerTest()
	fileRepository := repositories.NewFileRepository(nil)

	initEmailProcessHandlerTest()
	groupSettingIds := []uint{1, 2}
	attachments := []structs.Attachment{
		structs.Attachment{
			Filename: "test.jpg",
			FileType: "image/jpeg",
			Size:     0,
		},
		structs.Attachment{
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
		GroupSettingsIds: groupSettingIds,
	}

	for _, attachment := range attachments {
		testJpgBytes, err := fileRepository.GetTestJpgBytes()
		if err != nil {
			utils.PrintTestError(t, err, nil)
		}

		filePath := buildTempEmailFilePath(attachment.Filename)
		err = utils.WriteFile(filePath, testJpgBytes)
		if err != nil {
			utils.PrintTestError(t, err, nil)
		}

		defer os.Remove(filePath)
		defer os.Remove(buildTempEmailOcrFilePath(attachment.Filename))
	}

	metadataList := []structs.EmailMetadata{metadata}

	err := enqueueEmailProcessTasks(metadataList)
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	inspector, err := GetAsynqInspector()
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	taskInfos, err := inspector.ListPendingTasks(string(EmailReceiptProcessingQueue))
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if len(taskInfos) != 4 {
		utils.PrintTestError(t, len(taskInfos), 4)
	}

	for index, taskInfo := range taskInfos {
		var payload EmailProcessTaskPayload
		err := json.Unmarshal(taskInfo.Payload, &payload)
		if err != nil {
			utils.PrintTestError(t, err, nil)
		}

		expectedAttachment := attachments[index/2]
		expectedPayload := EmailProcessTaskPayload{
			GroupSettingsId: uint((index % 2) + 1),
			ImageForOcrPath: buildTempEmailOcrFilePath(expectedAttachment.Filename),
			TempFilePath:    buildTempEmailFilePath(expectedAttachment.Filename),
			Metadata:        metadata,
			Attachment:      expectedAttachment,
		}

		isEqual := reflect.DeepEqual(payload, expectedPayload)
		if !isEqual {
			utils.PrintTestError(t, payload, expectedPayload)
		}
	}
}

func teardownEmailProcessHandlerTest() {
	miniredisInstance.Close()
	repositories.ShutdownAsynqClient()
	teardown()
}
