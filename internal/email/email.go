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

func PollEmails() error {
	config := config.GetConfig()
	ticker := time.NewTicker(time.Duration(config.EmailPollingInterval) * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := CallClient("")
				if err != nil {
					logging.GetLogger().Println("Error polling emails")
					logging.GetLogger().Println(err.Error())
				}
			}
		}
	}()

	return nil
}

func CallClient(groupId string) error {
	logger := logging.GetLogger()
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)

	if groupId == "" {
		groupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ?", true)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
		err = pollEmailForGroupSettings(groupSettings)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
	} else {
		groupSettings, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ? AND group_id = ?", true, groupId)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
		err = pollEmailForGroupSettings(groupSettings)
		if err != nil {
			logger.Println(err.Error())
			return err
		}
	}
	return nil
}

func pollEmailForGroupSettings(groupSettings []models.GroupSettings) error {
	logger := logging.GetLogger()
	basePath := config.GetBasePath()

	bytesArr, err := json.Marshal(groupSettings)
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	var out bytes.Buffer
	cmd := exec.Command("python3", basePath+"/imap-client/client.py")
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(bytesArr)
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	var result []structs.EmailMetadata

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

func processEmails(emailMetadata []structs.EmailMetadata, groupSettings []models.GroupSettings) error {
	basePath := config.GetBasePath() + "/temp"
	db := repositories.GetDB()
	imagesToRemove := []string{}

	err := db.Transaction(func(tx *gorm.DB) error {
		receiptRepository := repositories.NewReceiptRepository(tx)
		receiptImageRepository := repositories.NewReceiptImageRepository(tx)

		for _, metadata := range emailMetadata {
			for _, attachment := range metadata.Attachments {
				path := basePath + "/" + attachment.Filename
				receipt, err := services.ReadReceiptImageFromFileOnly(path)
				if err != nil {
					return err
				}

				for _, groupSettingsId := range metadata.GroupSettingsIds {
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

					receipt.GroupId = groupSettingsToUse.GroupId
					receipt.Status = groupSettingsToUse.EmailDefaultReceiptStatus
					receipt.PaidByUserID = *groupSettingsToUse.EmailDefaultReceiptPaidById
					receipt.CreatedByString = "Email Integration"

					createdReceipt, err := receiptRepository.CreateReceipt(receipt, 0)
					if err != nil {
						return err
					}

					bytes, err := utils.ReadFile(path)
					if err != nil {
						return err
					}

					fileData := models.FileData{
						ReceiptId: createdReceipt.ID,
						Name:      attachment.Filename,
						FileType:  attachment.FileType,
						Size:      attachment.Size,
						ImageData: bytes,
					}

					_, err = receiptImageRepository.CreateReceiptImage(fileData)
					if err != nil {
						return err
					}

				}

				imagesToRemove = append(imagesToRemove, path)
			}
		}

		return nil
	})

	for _, path := range imagesToRemove {
		os.Remove(path)
	}

	if err != nil {
		logging.GetLogger().Println(err.Error())
		return err
	}

	return nil
}
