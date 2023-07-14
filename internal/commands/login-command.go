package commands

type LoginCommand struct {
	Username string `json:"userName"`
	Password string `json:"password"`
}
