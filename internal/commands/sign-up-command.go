package commands

type SignUpCommand struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Displayname string `json:"displayname"`
}
