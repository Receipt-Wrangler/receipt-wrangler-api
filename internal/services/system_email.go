package services

import (
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/commands"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/utils"
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

func (service SystemEmailService) CheckEmailConnectivity(command commands.CheckEmailConnectivityCommand) error {
	if command.ID > 0 {
		stringId := simpleutils.UintToString(command.ID)
		systemEmailRepository := repositories.NewSystemEmailRepository(nil)
		systemEmail, err := systemEmailRepository.GetSystemEmailById(stringId)
		if err != nil {
			return err
		}

		command.Host = systemEmail.Host
		command.Port = systemEmail.Port
		command.Username = systemEmail.Username

		cleartextPassword, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
		if err != nil {
			return err
		}

		command.Password = cleartextPassword
	}

	// TODO: call python script

	return nil
}
