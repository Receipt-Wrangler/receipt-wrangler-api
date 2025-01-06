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

	if systemSettings.AsynqEmailPollingId != "" {
		err = UnregisterTask(systemSettings.AsynqEmailPollingId)
		if err != nil {
			return err
		}
	}

	task := asynq.NewTask(EmailPoll, nil)
	entryId, err := RegisterTask(GetPollTimeString(systemSettings.EmailPollingInterval), task)
	if err != nil {
		return err
	}

	_, err = systemSettingsRepository.UpdateAsynqEmailPollingId(entryId)
	if err != nil {
		return err
	}

	return nil
}

func GetPollTimeString(pollingInterval int) string {
	return fmt.Sprintf("every %ds", pollingInterval)
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

	// TOOD: kick off processing task for one email by iterating over metadata
	err = enqueueEmailProcessTasks(result, groupSettings)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	return nil
}

func enqueueEmailProcessTasks(metadataList []structs.EmailMetadata, groupSettings []models.GroupSettings) error {
	basePath := config.GetBasePath() + "/temp"
	fileRepository := repositories.NewCategoryRepository(nil)

	for _, metadata := range metadataList {

		for _, attachment := range metadata.Attachments {
			tempFilePath := basePath + "/" + attachment.Filename
			defer os.Remove(tempFilePath)

			imageForOcrPath := basePath + "/" + "image-" + attachment.Filename
			defer os.Remove(imageForOcrPath)

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
				_, err = EnqueueTask(task)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
