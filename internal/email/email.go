package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"gorm.io/gorm"
)

var ticker *time.Ticker

func StartEmailPolling() error {
	if ticker != nil {
		ticker.Stop()
	}

	err := PollEmails()
	if err != nil {
		return err
	}

	return nil
}

func PollEmails() error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	ticker = time.NewTicker(time.Duration(systemSettings.EmailPollingInterval) * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := CallClient(true, nil)
				if err != nil {
					logging.GetLogger().Println("Error polling emails")
					logging.GetLogger().Println(err.Error())
				}
			}
		}
	}()

	return nil
}

func CallClient(pollAllGroups bool, groupIds []string) error {
	logger := logging.GetLogger()
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	var groupSettings []models.GroupSettings

	if pollAllGroups {
		allGroupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ?", true)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
		groupSettings = allGroupSettings
	} else {
		someGroupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ? AND group_id IN ?", true, groupIds)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
		groupSettings = someGroupSettings
	}

	err := pollEmailForGroupSettings(groupSettings)
	if err != nil {
		logger.Println(err.Error())
		return err
	}
	return nil
}

func pollEmailForGroupSettings(groupSettings []models.GroupSettings) error {
	logger := logging.GetLogger()
	basePath := config.GetBasePath()
	groupSettingsWithPassword := make([]models.GroupSettingsWithSystemEmailPassword, len(groupSettings))

	// TODO: Could be more efficient by only decrypting the passwords once for each email
	for i := range groupSettings {
		cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), groupSettings[i].SystemEmail.Password)
		if err != nil {
			logger.Println(err.Error())
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
		logger.Println(err.Error())
		return err
	}

	var out bytes.Buffer
	cmd := exec.Command("python3", basePath+"/imap-client/client_new.py")
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(bytesArr)
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	// TODO: fix unmarshal
	var result map[structs.ComparableEmailMetadata][]uint

	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	logger.Println("Emails metadata captured: ", result)

	err = processEmails(result, groupSettings)
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	return nil
}

func processEmails(metadataMap map[structs.ComparableEmailMetadata][]uint, groupSettings []models.GroupSettings) error {
	basePath := config.GetBasePath() + "/temp"
	db := repositories.GetDB()
	fileRepository := repositories.NewCategoryRepository(nil)

	for metadata, groupSettingIds := range metadataMap {
		var attachments []structs.Attachment

		err := json.Unmarshal([]byte(metadata.Attachments), &attachments)
		if err != nil {
			return err
		}

		for _, attachment := range attachments {
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

			baseCommand, _, err := services.ReadReceiptImageFromFileOnly(imageForOcrPath)
			if err != nil {
				return err
			}

			for _, groupSettingsId := range groupSettingIds {
				groupSettingsToUse := models.GroupSettings{}

				for _, groupSetting := range groupSettings {
					if groupSetting.ID == groupSettingsId {
						groupSettingsToUse = groupSetting
						break
					}
				}

				if groupSettingsToUse.ID == 0 {
					return fmt.Errorf("could not find group settings with id %d", groupSettingsId)
				}

				command := baseCommand
				command.GroupId = groupSettingsToUse.GroupId
				command.Status = groupSettingsToUse.EmailDefaultReceiptStatus
				command.PaidByUserID = *groupSettingsToUse.EmailDefaultReceiptPaidById
				command.CreatedByString = "Email Integration"

				err = db.Transaction(func(tx *gorm.DB) error {
					receiptRepository := repositories.NewReceiptRepository(tx)
					receiptImageRepository := repositories.NewReceiptImageRepository(tx)

					createdReceipt, err := receiptRepository.CreateReceipt(command, 0)
					if err != nil {
						return err
					}

					fileData := models.FileData{
						ReceiptId: createdReceipt.ID,
						Name:      attachment.Filename,
						FileType:  attachment.FileType,
						Size:      attachment.Size,
					}

					_, err = receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
					if err != nil {
						return err
					}

					return nil
				})
			}
		}
	}

	return nil
}
