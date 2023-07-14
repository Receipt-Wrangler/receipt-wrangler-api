package commands

type LoginCommand struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
