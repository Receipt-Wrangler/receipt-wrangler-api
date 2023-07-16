package commands

import "receipt-wrangler/api/internal/models"

type SignUpCommand struct {
	Username    string          `json:"username"`
	Password    string          `json:"password"`
	DisplayName string          `json:"displayName"`
	IsDummyUser bool            `json:"isDummyUser"`
	UserRole    models.UserRole `json:"userRole"`
}
