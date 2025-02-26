package wranglerasynq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"os"
	"os/exec"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func StartEmailPolling() error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	inspector, err := GetAsynqInspector()
	if err != nil {
		return nil
	}
	defer inspector.Close()

	_, err = inspector.DeleteAllScheduledTasks(string(models.EmailPollingQueue))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_INFO, err.Error())
	}

	_, err = inspector.DeleteAllScheduledTasks(string(models.EmailReceiptImageCleanupQueue))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_INFO, err.Error())
	}

	payload := EmailPollTaskPayload{
		PollAllGroups: true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pollTask := asynq.NewTask(EmailPoll, payloadBytes)
	_, err = RegisterTask(GetPollTimeString(systemSettings.EmailPollingInterval), pollTask, models.EmailPollingQueue, 0)
	if err != nil {
		return err
	}

	cleanUpTask := asynq.NewTask(EmailProcessImageCleanUp, nil)
	_, err = RegisterTask(GetPollTimeString(systemSettings.EmailPollingInterval*2), cleanUpTask, models.EmailReceiptImageCleanupQueue, 0)
	if err != nil {
		return err
	}

	return nil
}

func GetPollTimeString(pollingInterval int) string {
	return fmt.Sprintf("@every %ds", pollingInterval)
}

func CallClient(pollAllGroups bool, groupIds []string) error {
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	var groupSettings []models.GroupSettings

	if pollAllGroups {
		allGroupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ?", true)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			return err
		}
		groupSettings = allGroupSettings
	} else {
		someGroupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ? AND group_id IN ?", true, groupIds)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			return err
		}
		groupSettings = someGroupSettings
	}

	if len(groupSettings) == 0 {
		logging.LogStd(logging.LOG_LEVEL_INFO, "No group settings enabled for email polling")
		return nil
	}

	err := pollEmailForGroupSettings(groupSettings)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}
	return nil
}

func pollEmailForGroupSettings(groupSettings []models.GroupSettings) error {
	basePath := config.GetBasePath()
	groupSettingsWithPassword := make([]models.GroupSettingsWithSystemEmailPassword, len(groupSettings))

	// TODO: Could be more efficient by only decrypting the passwords once for each email
	for i := range groupSettings {
		cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), groupSettings[i].SystemEmail.Password)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			return err
		}

		var groupSettingWithPassword models.GroupSettingsWithSystemEmailPassword
		groupSettingWithPassword.BaseModel = groupSettings[i].BaseModel
		groupSettingWithPassword.GroupSettings = groupSettings[i]
		groupSettingWithPassword.SystemEmail = models.SystemEmailWithPassword{
			BaseModel:   groupSettings[i].SystemEmail.BaseModel,
			SystemEmail: groupSettings[i].SystemEmail,
			Password:    cleartextPassword,
		}

		groupSettingsWithPassword[i] = groupSettingWithPassword
	}

	bytesArr, err := json.Marshal(groupSettingsWithPassword)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	var out bytes.Buffer
	cmd := exec.Command("python3", basePath+"/imap-client/client.py")
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(bytesArr)
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	var result []structs.EmailMetadata

	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	logging.LogStd(logging.LOG_LEVEL_INFO, "Emails metadata captured: ", result)
	err = enqueueEmailProcessTasks(result)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	return nil
}

func enqueueEmailProcessTasks(metadataList []structs.EmailMetadata) error {
	fileRepository := repositories.NewFileRepository(nil)

	for _, metadata := range metadataList {

		for _, attachment := range metadata.Attachments {
			tempFilePath := buildTempEmailFilePath(attachment.Filename)
			imageForOcrPath := buildTempEmailOcrFilePath(attachment.Filename)

			fileBytes, err := utils.ReadFile(tempFilePath)
			if err != nil {
				return err
			}

			ocrBytes, err := fileRepository.GetBytesFromImageBytes(fileBytes)
			if err != nil {
				return err
			}

			err = utils.WriteFile(imageForOcrPath, ocrBytes)
			if err != nil {
				return err
			}

			for _, groupSettingsId := range metadata.GroupSettingsIds {
				payload := EmailProcessTaskPayload{
					GroupSettingsId: groupSettingsId,
					ImageForOcrPath: imageForOcrPath,
					TempFilePath:    tempFilePath,
					Metadata:        metadata,
					Attachment:      attachment,
				}
				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					return err
				}

				task := asynq.NewTask(EmailProcess, payloadBytes)
				_, err = EnqueueTask(task, models.EmailReceiptProcessingQueue)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func buildTempEmailFilePath(attachmentFileName string) string {
	fileRepository := repositories.NewFileRepository(nil)
	return fileRepository.GetTempDirectoryPath() + "/" + attachmentFileName
}

func buildTempEmailOcrFilePath(attachmentFileName string) string {
	fileRepository := repositories.NewFileRepository(nil)
	return fileRepository.GetTempDirectoryPath() + "/" + "image-" + attachmentFileName
}
