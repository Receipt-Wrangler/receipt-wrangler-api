package commands

import "receipt-wrangler/api/internal/models"

func GetDefaultAdminSignUpCommand() SignUpCommand {
	return SignUpCommand{
		Username:    "admin",
		DisplayName: "Admin",
		Password:    "admin",
		IsDummyUser: false,
		UserRole:    models.ADMIN,
	}
}
