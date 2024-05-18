package services

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"
	"time"
)

type SystemEmailService struct {
	BaseService
}

func NewSystemEmailService(tx *gorm.DB) SystemEmailService {
	service := SystemEmailService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

func (service SystemEmailService) CheckEmailConnectivity(command commands.CheckEmailConnectivityCommand, userId uint) (models.SystemTask, error) {
	hostIsEmpty := len(command.Host) == 0
	portIsEmpty := len(command.Port) == 0
	usernameIsEmpty := len(command.Username) == 0
	passwordIsEmpty := len(command.Password) == 0

	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:                 models.SYSTEM_EMAIL_CONNECTIVITY_CHECK,
		Status:               models.SYSTEM_TASK_FAILED,
		AssociatedEntityType: models.SYSTEM_EMAIL,
		AssociatedEntityId:   command.ID,
		RanByUserId:          &userId,
	}

	if command.ID > 0 && hostIsEmpty && portIsEmpty && usernameIsEmpty && passwordIsEmpty {
		stringId := simpleutils.UintToString(command.ID)
		systemEmailRepository := repositories.NewSystemEmailRepository(nil)
		systemEmail, err := systemEmailRepository.GetSystemEmailById(stringId)
		if err != nil {
			return models.SystemTask{}, err
		}

		command.Host = systemEmail.Host
		command.Port = systemEmail.Port
		command.Username = systemEmail.Username

		cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
		if err != nil {
			return models.SystemTask{}, err
		}

		command.Password = cleartextPassword
	}

	commandBytes, err := json.Marshal(command)
	if err != nil {
		return models.SystemTask{}, err
	}

	basePath := config.GetBasePath()
	path := basePath + "/imap-client/check_connection.py"

	var out bytes.Buffer

	cmd := exec.Command("python3", path)
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(commandBytes)
	cmd.Env = os.Environ()

	systemTaskCommand.StartedAt = time.Now()

	// Note: We ignore error from cmd.Run to capture the output in the system task, so resulting HTTP status code is always 200
	err = cmd.Run()
	if err != nil {
		systemTaskCommand.ResultDescription = err.Error()
	} else {
		systemTaskCommand.Status = models.SYSTEM_TASK_SUCCEEDED
	}
	err = nil

	now := time.Now()
	systemTaskCommand.EndedAt = &now

	if command.ID > 0 {
		return repositories.NewSystemTaskRepository(nil).CreateSystemTask(systemTaskCommand)
	}

	return models.SystemTask{
		Status: systemTaskCommand.Status,
	}, err
}
