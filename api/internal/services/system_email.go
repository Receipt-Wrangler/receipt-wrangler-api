package services

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/emailclient"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"time"
)

// checkEmailConnection logs in to the mailbox and selects INBOX. A variable so
// tests can stand in for the IMAP server.
var checkEmailConnection = emailclient.Connector{}.CheckConnection

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
		stringId := utils.UintToString(command.ID)
		systemEmailRepository := repositories.NewSystemEmailRepository(nil)
		systemEmail, err := systemEmailRepository.GetSystemEmailById(stringId)
		if err != nil {
			return models.SystemTask{}, err
		}

		command.Host = systemEmail.Host
		command.Port = systemEmail.Port
		command.Username = systemEmail.Username
		command.UseStartTLS = systemEmail.UseStartTLS

		cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
		if err != nil {
			return models.SystemTask{}, err
		}

		command.Password = cleartextPassword
	}

	systemTaskCommand.StartedAt = time.Now()

	// A failed check is still a 200: the outcome is reported on the system task.
	err := checkEmailConnection(emailclient.Credentials{
		Host:        command.Host,
		Port:        command.Port,
		Username:    command.Username,
		Password:    command.Password,
		UseStartTLS: command.UseStartTLS,
	})
	if err != nil {
		systemTaskCommand.ResultDescription = err.Error()
	} else {
		systemTaskCommand.Status = models.SYSTEM_TASK_SUCCEEDED
		systemTaskCommand.ResultDescription = "Connection successful"
	}

	now := time.Now()
	systemTaskCommand.EndedAt = &now

	if command.ID > 0 {
		return repositories.NewSystemTaskRepository(nil).CreateSystemTask(systemTaskCommand)
	}

	return models.SystemTask{
		Status:            systemTaskCommand.Status,
		ResultDescription: systemTaskCommand.ResultDescription,
	}, nil
}
