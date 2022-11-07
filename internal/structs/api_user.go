package structs

type APIUser struct {
	ID          uint   `json:"id"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}
