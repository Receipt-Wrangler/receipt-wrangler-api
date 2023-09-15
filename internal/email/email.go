package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/repositories"
)

func PollEmails() error {
	err := callClient()
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func callClient() error {
	logger := logging.GetLogger()
	basePath := config.GetBasePath()
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	groupSettings, err := groupSettingsRepository.GetAllGroupSettings()
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	bytesArr, err := json.Marshal(groupSettings)
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	var out bytes.Buffer
	cmd := exec.Command("python3", basePath+"/imap-client/client.py")
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(bytesArr)

	err = cmd.Run()
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	var result interface{}

	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	fmt.Println(result)

	err = processEmails()
	if err != nil {
		logger.Println(err.Error())
		return err
	}

	return nil
}

func processEmails() error {
	return nil
}
