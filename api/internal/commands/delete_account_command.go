package commands

// DeleteAccountCommand confirms a self-delete. An account with a password of its
// own confirms with Password. An account without one -- created by OIDC
// provisioning -- confirms with ConfirmationCode, which signing in at one of its
// linked providers again hands back.
type DeleteAccountCommand struct {
	Password         string `json:"password"`
	ConfirmationCode string `json:"confirmationCode"`
}
